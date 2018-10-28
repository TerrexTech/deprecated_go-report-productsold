package main

import (
	"os"

	"github.com/TerrexTech/go-authserver-query/kafka"
	"github.com/TerrexTech/go-commonutils/commonutil"
)

// initKafkaIOLogin creates a KafkaIO from KafkaAdapter based on set environment variables.
func initKafkaIOReport() (*kafka.IO, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	consumerGroupName := os.Getenv("KAFKA_CONSUMER_GROUP_REPORT")
	consumerTopics := os.Getenv("KAFKA_CONSUMER_TOPIC_REPORT")
	responseTopic := os.Getenv("KAFKA_PRODUCER_TOPIC_REPORT")

	kafkaAdapter := &kafka.Adapter{
		Brokers:           *commonutil.ParseHosts(brokers),
		ConsumerGroupName: consumerGroupName,
		ConsumerTopics:    *commonutil.ParseHosts(consumerTopics),
		ProducerTopic:     responseTopic,
	}

	return kafkaAdapter.InitIO()
}
