package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"credit-authorization-ledger/internal/authorization"
	"credit-authorization-ledger/internal/config"
	"credit-authorization-ledger/internal/database"
	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/internal/tracing"

	"go.opentelemetry.io/otel"
)

func main() {
	cfg := config.Load()
	tracer := tracing.InitTracer("authorization-service")
	otel.SetTracerProvider(tracer)

	db, err := database.NewPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Run database migrations for the authorization service
	database.RunMigrations(db, "internal/authorization/migrations")

	authService := authorization.NewService(db)

	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, "authorization-group")
	defer kafkaConsumer.Close()

	// Subscribe to the topic where authorization requests are sent
	go kafkaConsumer.Consume([]string{"authorization-requests"}, authService.HandleAuthorizationRequest)

	log.Println("Authorization service started...")

	// Graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Authorization service shutting down...")
}
