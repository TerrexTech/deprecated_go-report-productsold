package report

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/TerrexTech/go-commonutils/commonutil"
	mongo "github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/TerrexTech/uuuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flash sale Suite")
}

// newTimeoutContext creates a new WithTimeout context with specified timeout.
func newTimeoutContext(timeout uint32) (ctx.Context, ctx.CancelFunc) {
	return ctx.WithTimeout(
		ctx.Background(),
		time.Duration(timeout)*time.Millisecond,
	)
}

type Env struct {
	Flashdb     DBI
	Metricdb    DBI
	Inventorydb DBI
}

type mockDb struct {
	collection *mongo.Collection
}

var _ = Describe("Mongo service test", func() {
	var (
		// jsonString string
		// mgTable         *mongo.Collection
		// client          mongo.Client
		resourceTimeout uint32
		testDatabase    string
		// clientConfig    mongo.ClientConfig
		// c               *mongo.Collection
		// dataCount int
		// unixTime        int64
		configInv DBIConfig
		// configMetric DBIConfig
		// configInv    DBIConfig
	)

	testDatabase = "rns_report_test"
	resourceTimeout = 3000
	// dataCount = 5

	dropTestDatabase := func(dbConfig DBIConfig) {
		clientConfig := mongo.ClientConfig{
			Hosts:               dbConfig.Hosts,
			Username:            dbConfig.Username,
			Password:            dbConfig.Password,
			TimeoutMilliseconds: dbConfig.TimeoutMilliseconds,
		}
		client, err := mongo.NewClient(clientConfig)
		Expect(err).ToNot(HaveOccurred())

		dbCtx, dbCancel := newTimeoutContext(resourceTimeout)
		err = client.Database(testDatabase).Drop(dbCtx)
		dbCancel()
		Expect(err).ToNot(HaveOccurred())

		err = client.Disconnect()
		Expect(err).ToNot(HaveOccurred())
	}

	GenerateTestDB := func(dbConfig DBIConfig, schema interface{}) (*DB, error) {
		config := mongo.ClientConfig{
			Hosts:               dbConfig.Hosts,
			Username:            dbConfig.Username,
			Password:            dbConfig.Password,
			TimeoutMilliseconds: dbConfig.TimeoutMilliseconds,
		}

		client, err := mongo.NewClient(config)
		if err != nil {
			err = errors.Wrap(err, "Error creating DB-client")
			return nil, err
		}

		conn := &mongo.ConnectionConfig{
			Client:  client,
			Timeout: 5000,
		}

		// ====> Create New Collection
		collConfig := &mongo.Collection{
			Connection:   conn,
			Database:     dbConfig.Database,
			Name:         dbConfig.Collection,
			SchemaStruct: schema,
			// Indexes:      indexConfigs,
		}
		c, err := mongo.EnsureCollection(collConfig)
		if err != nil {
			err = errors.Wrap(err, "Error creating DB-client")
			return nil, err
		}
		return &DB{
			collection: c,
		}, nil
	}

	CreateNewUUID := func() (uuuid.UUID, error) {
		uuid, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		return uuid, nil
	}

	BeforeEach(func() {
		configInv = DBIConfig{
			Hosts:               *commonutil.ParseHosts("localhost:27017"),
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 3000,
			Database:            testDatabase,
			Collection:          "agg_inventory",
		}

		// configMetric = DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            testDatabase,
		// 	Collection:          "agg_metric",
		// }

		// configInv = DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            testDatabase,
		// 	Collection:          "agg_inventory",
		// }
	})

	AfterEach(func() {
		configInv := DBIConfig{
			Hosts:               *commonutil.ParseHosts("localhost:27017"),
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 3000,
			Database:            "rns_projections",
			Collection:          "agg_inventory",
		}

		// configMetric := DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            "rns_projections",
		// 	Collection:          "agg_metric",
		// }

		// configInv := DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            "rns_projections",
		// 	Collection:          "agg_inventory",
		// }
		dropTestDatabase(configInv)
		// err := client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())

		// dropTestDatabase(configMetric)
		// err = client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())

		// dropTestDatabase(configInv)
		// err = client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())
	})

	It("Search inventory when time fields are missing", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		// _, err = GenerateTestDB(configMetric, &Metric{})
		// Expect(err).ToNot(HaveOccurred())

		// _, err = GenerateTestDB(configInv, &Inventory{})
		// Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"field": "sku", "type":"int","equal":"343434"}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		searchResults, err := dbInventory.InvAdvSearch(query)
		Expect(err).ToNot(HaveOccurred())

		log.Println(searchResults)

		for _, v := range searchResults {
			Expect(v.Name).To(Equal("test"))
		}
	})

	It("Should give error if type and fields are empty", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"upper_limit": 4000, "lower_limit":3000}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.InvAdvSearch(query)
		Expect(err).To(HaveOccurred())
	})

	It("Should give error if field is empty", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"type":"int","equal":"343434","upper_limit": 4000, "lower_limit":3000}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.InvAdvSearch(query)
		Expect(err).To(HaveOccurred())
	})

	It("Should give error if type is empty", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"field":"sku","equal":"343434","upper_limit": 4000, "lower_limit":3000}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.InvAdvSearch(query)
		Expect(err).To(HaveOccurred())
	})

	It("Should give error if equal is empty", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"field":"sku","type":"int","upper_limit": 4000, "lower_limit":3000}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.InvAdvSearch(query)
		Expect(err).To(HaveOccurred())
	})

	It("Should give error if field is empty", func() {

		dbInventory, err := GenerateTestDB(configInv, &Inventory{})
		Expect(err).ToNot(HaveOccurred())

		rscustomerId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		invData := fmt.Sprintf(`{"item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","total_weight":2000,"price":100,"lot":"A-1","date_arrived":3000,"expiry_date":4000,"timestamp":3500,"rs_customer_id":"%v","waste_weight":10,"donate_weight":15,"date_sold":3600,"sale_price": 18,"sold_weight":800,"prod_quantity":50}`, itemId, deviceId, rscustomerId)

		inv := Inventory{}

		err = json.Unmarshal([]byte(invData), &inv)
		Expect(err).ToNot(HaveOccurred())

		_, err = dbInventory.collection.InsertOne(inv)
		Expect(err).ToNot(HaveOccurred())

		spData := fmt.Sprintf(`{"inventory":[{"field":"sku","type":"int","lower_limit":3000, "upper_limit": 4000}]}`)

		var query map[string][]SearchParam

		err = json.Unmarshal([]byte(spData), &query)
		Expect(err).ToNot(HaveOccurred())

		searchResults, err := dbInventory.InvAdvSearch(query)
		Expect(err).ToNot(HaveOccurred())

		log.Println(searchResults)

		for _, v := range searchResults {
			Expect(v.Name).To(Equal("test"))
		}
	})

})
