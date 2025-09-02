package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	PostgresURL  string
	DynamoDBURL  string
	KafkaBrokers []string
	AWSRegion    string
	SQSQueueURL  string
}

func Load() *Config {
	return &Config{
		PostgresURL:  getEnv("POSTGRES_URL", "postgres://user:password@postgres:5432/credit_ledger?sslmode=disable"),
		DynamoDBURL:  getEnv("DYNAMODB_URL", "http://dynamodb:8000"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ","),
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		SQSQueueURL:  getEnv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/my-queue"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using fallback for %s", key)
	return fallback
}