package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"credit-authorization-ledger/internal/config"
	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/internal/saga"
	"credit-authorization-ledger/internal/tracing"

	"go.opentelemetry.io/otel"
)

func main() {
	cfg := config.Load()
	tracer := tracing.InitTracer("saga-orchestrator")
	otel.SetTracerProvider(tracer)

	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}

	orchestrator := saga.NewOrchestrator(producer)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, "saga-orchestrator-group")

	// The orchestrator listens to the initial request and the outcomes of each step
	topics := []string{
		"credit-authorization-requested", // Initial trigger
		"authorization-succeeded",
		"authorization-failed",
		"ledger-update-succeeded",
		"ledger-update-failed",
	}
	go consumer.Consume(topics, orchestrator.HandleMessage)

	log.Println("SAGA Orchestrator started...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("SAGA Orchestrator shutting down...")
}