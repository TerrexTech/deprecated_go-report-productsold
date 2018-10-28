package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/TerrexTech/go-commonutils/commonutil"
	"github.com/TerrexTech/go-eventspoll/poll"
	esmodel "github.com/TerrexTech/go-eventstore-models/model"
	"github.com/TerrexTech/go-report-productsold/report"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type Env struct {
	Flashdb     report.DBI
	Metricdb    report.DBI
	Inventorydb report.DBI
}

type KaRespData struct {
	SKU         int64
	Name        string
	TotalWeight float64
	SoldWeight  float64
	Price       float64
}

func main() {
	// Load environment-file.
	// Env vars will be read directly from environment if this file fails loading
	err := godotenv.Load()
	if err != nil {
		err = errors.Wrap(err,
			".env file not found, env-vars will be read as set in environment",
		)
		log.Println(err)
	}

	missingVar, err := commonutil.ValidateEnv(
		"KAFKA_BROKERS",
		"KAFKA_CONSUMER_EVENT_GROUP",
		"KAFKA_CONSUMER_EVENT_TOPIC",
		"KAFKA_CONSUMER_EVENT_QUERY_GROUP",
		"KAFKA_CONSUMER_EVENT_QUERY_TOPIC",
		"KAFKA_PRODUCER_EVENT_QUERY_TOPIC",
		"KAFKA_PRODUCER_RESPONSE_TOPIC",

		"MONGO_HOSTS",
		"MONGO_USERNAME",
		"MONGO_PASSWORD",
		"MONGO_DATABASE",
		// "MONGO_CONNECTION_TIMEOUT_MS",
		// "MONGO_RESOURCE_TIMEOUT_MS",
	)
	if err != nil {
		log.Fatalf(
			"Error: Environment variable %s is required but was not found", missingVar,
		)
	}

	hosts := os.Getenv("MONGO_HOSTS")
	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")
	database := os.Getenv("MONGO_DATABASE")
	collectionFlash := os.Getenv("MONGO_FLASH_COLLECTION")
	collectionInv := os.Getenv("MONGO_INV_COLLECTION")
	collectionMet := os.Getenv("MONGO_METRIC_COLLECTION")

	consumerEventgroup := os.Getenv("KAFKA_CONSUMER_EVENT_GROUP")
	consumerEventQueryGroup := os.Getenv("KAFKA_CONSUMER_EVENT_QUERY_GROUP")
	consumerEventTopic := os.Getenv("KAFKA_CONSUMER_EVENT_TOPIC")
	consumerEventQueryTopic := os.Getenv("KAFKA_CONSUMER_EVENT_QUERY_TOPIC")
	producerEventQueryTopic := os.Getenv("KAFKA_PRODUCER_EVENT_QUERY_TOPIC")
	producerResponseTopic := os.Getenv("KAFKA_PRODUCER_RESPONSE_TOPIC")

	timeoutMilliStr := os.Getenv("MONGO_TIMEOUT")
	parsedTimeoutMilli, err := strconv.Atoi(timeoutMilliStr)
	if err != nil {
		err = errors.Wrap(err, "Error converting Timeout value to int32")
		log.Println(err)
		log.Println("MONGO_TIMEOUT value will be set to 3000 as default value")
		parsedTimeoutMilli = 3000
	}
	timeoutMilli := uint32(parsedTimeoutMilli)

	log.Println(hosts)
	configFlash := report.DBIConfig{
		Hosts:               *commonutil.ParseHosts(hosts),
		Username:            username,
		Password:            password,
		TimeoutMilliseconds: timeoutMilli,
		Database:            database,
		Collection:          collectionFlash,
	}

	configMetric := report.DBIConfig{
		Hosts:               *commonutil.ParseHosts(hosts),
		Username:            username,
		Password:            password,
		TimeoutMilliseconds: timeoutMilli,
		Database:            database,
		Collection:          collectionMet,
	}

	configInv := report.DBIConfig{
		Hosts:               *commonutil.ParseHosts(hosts),
		Username:            username,
		Password:            password,
		TimeoutMilliseconds: timeoutMilli,
		Database:            database,
		Collection:          collectionInv,
	}

	dbFlash, err := report.GenerateDB(configFlash, &report.Flash{})
	if err != nil {
		err = errors.Wrap(err, "Error connecting to Inventory DB")
		log.Println(err)
		return
	}

	log.Println(configInv, configMetric)

	dbMetric, err := report.GenerateDB(configMetric, &report.Metric{})
	if err != nil {
		err = errors.Wrap(err, "Error connecting to Inventory DB")
		log.Println(err)
		return
	}

	dbInventory, err := report.GenerateDB(configInv, &report.Inventory{})
	if err != nil {
		err = errors.Wrap(err, "Error connecting to Inventory DB")
		log.Println(err)
		return
	}

	// This Env is in file route_handlers.go
	env := &Env{
		Flashdb:     dbFlash,
		Metricdb:    dbMetric,
		Inventorydb: dbInventory,
	}

	kc := poll.KafkaConfig{
		Brokers: []string{"kafka:9092"},

		ConsumerEventGroup:      consumerEventgroup,
		ConsumerEventQueryGroup: consumerEventQueryGroup,

		ConsumerEventTopic:      consumerEventTopic,
		ConsumerEventQueryTopic: consumerEventQueryTopic,
		ProducerEventQueryTopic: producerEventQueryTopic,
		ProducerResponseTopic:   producerResponseTopic,
	}
	ioConfig := poll.IOConfig{
		AggregateID: 4,
		// Choose what type of events we need process
		// Remember, adding a type here and not processing/listening to it will cause deadlocks!
		ReadConfig: poll.ReadConfig{
			EnableQuery: true,
		},
		KafkaConfig:     kc,
		MongoCollection: dbInventory.Collection(),
		// The number of times we allow failure to get max aggregate-version from DB.
		// Check docs for more info.
		MongoFailThreshold: 300,
	}

	eventPoll, err := poll.Init(ioConfig)
	if err != nil {
		err = errors.Wrap(err, "Error creating EventPoll service")
		log.Fatalln(err)
	}

	// Handle Insert events
	for eventResp := range eventPoll.Query() {
		go func() {
			kafkaResp := handleQuery(eventResp, env)
			if kafkaResp != nil {
				eventPoll.ProduceResult() <- kafkaResp
			}
		}()
	}
}

func handleQuery(eventResp *poll.EventResponse, env *Env) *esmodel.KafkaResponse {
	log.Printf("%+v", eventResp)
	err := eventResp.Error
	if err != nil {
		err = errors.Wrap(err, "Some error occurred")
		log.Println(err)
		return nil
	}

	var sParam map[string][]report.SearchParam
	event := eventResp.Event
	data := event.Data
	err = json.Unmarshal(data, &sParam)
	if err != nil {
		err = errors.Wrap(err, "Some error occurred")
		log.Println(err)
		return nil
	}

	searchResults, err := env.Inventorydb.InvAdvSearch(sParam)
	if err != nil {
		err = errors.Wrap(err, "Unable to search inventory using search parameters")
		log.Println(err)
		return nil
	}

	var kaResp []KaRespData

	for _, v := range searchResults {
		kaResp = append(kaResp, KaRespData{
			SKU:         v.SKU,
			Name:        v.Name,
			TotalWeight: v.TotalWeight,
			SoldWeight:  v.SoldWeight,
			Price:       v.Price,
		})
	}

	kaRespByte, err := json.Marshal(&kaResp)
	if err != nil {
		err = errors.Wrap(err, "Did not marshal KaRespData")
		log.Println(err)
		return nil
	}

	return &esmodel.KafkaResponse{
		AggregateID:   event.AggregateID,
		CorrelationID: event.CorrelationID,
		Result:        kaRespByte,
	}

}
