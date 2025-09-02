
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"credit-authorization-ledger/internal/config"
	"credit-authorization-ledger/internal/database"
	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/internal/ledger"
	"credit-authorization-ledger/internal/tracing"

	"go.opentelemetry.io/otel"
)

func main() {
	cfg := config.Load()
	tracer := tracing.InitTracer("ledger-service")
	otel.SetTracerProvider(tracer)

	db, err := database.NewPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Run database migrations for the ledger service
	database.RunMigrations(db, "internal/ledger/migrations")

	ledgerService := ledger.NewService(db)

	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, "ledger-group")
	defer kafkaConsumer.Close()

	// Subscribe to the topic where successful authorizations are published
	go kafkaConsumer.Consume([]string{"authorization-succeeded"}, ledgerService.HandleCreditLedger)

	log.Println("Ledger service started...")

	// Graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Ledger service shutting down...")
}