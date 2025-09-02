package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"credit-authorization-ledger/internal/config"
	"credit-authorization-ledger/internal/database"
	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/internal/outbox"
	"credit-authorization-ledger/internal/tracing"
)

func main() {
	cfg := config.Load()
	tracing.InitTracer("outbox-processor")

	db, err := database.NewPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// The outbox table migration is run by other services, but it could be run here too for standalone safety.
	// database.RunMigrations(db, "internal/database/migrations")

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	processor := outbox.NewProcessor(db, kafkaProducer)

	// Use a ticker to periodically process messages
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Outbox processor started...")

	for {
		select {
		case <-ticker.C:
			if err := processor.ProcessOutboxMessages(); err != nil {
				log.Printf("error processing outbox messages: %v", err)
			}
		case <-shutdown:
			log.Println("Outbox processor shutting down...")
			return
		}
	}
}