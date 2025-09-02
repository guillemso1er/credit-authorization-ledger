package saga

import (
	"context"
	"encoding/json"
	"log"

	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/pkg/events"

	k "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type Orchestrator struct {
	producer *kafka.Producer
}

func NewOrchestrator(p *kafka.Producer) *Orchestrator {
	return &Orchestrator{producer: p}
}

func (o *Orchestrator) HandleMessage(ctx context.Context, msg k.Message) error {
	tr := otel.Tracer("saga-orchestrator")
	ctx, span := tr.Start(ctx, "HandleSagaMessage")
	defer span.End()

	log.Printf("Orchestrator received message on topic: %s", msg.Topic)

	switch msg.Topic {
	case "credit-authorization-requested":
		return o.startAuthorization(ctx, msg)
	case "authorization-succeeded":
		return o.startLedgerUpdate(ctx, msg)
	case "ledger-update-succeeded":
		log.Printf("SAGA completed successfully for transaction: %s", string(msg.Key))
		// Optionally publish a "SagaCompleted" event
	case "authorization-failed":
		fallthrough
	case "ledger-update-failed":
		log.Printf("SAGA failed for transaction: %s. Starting compensation.", string(msg.Key))
		// Here you would trigger compensating transactions
	}
	return nil
}

func (o *Orchestrator) startAuthorization(ctx context.Context, msg k.Message) error {
	// The API gateway should have sent this. We forward it to the auth service.
	return o.producer.Publish(ctx, "authorization-requests", string(msg.Key), msg.Value)
}

func (o *Orchestrator) startLedgerUpdate(ctx context.Context, msg k.Message) error {
	var authEvent events.AuthorizationSucceeded
	if err := json.Unmarshal(msg.Value, &authEvent); err != nil {
		return err
	}
	// The authorization service succeeded. Now we tell the ledger service.
	// The message is already the correct one, so we just forward it.
	return o.producer.Publish(ctx, "authorization-succeeded", authEvent.TransactionID, msg.Value)
}