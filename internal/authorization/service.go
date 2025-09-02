package authorization

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"credit-authorization-ledger/internal/outbox"
	"credit-authorization-ledger/pkg/events"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// HandleAuthorizationRequest processes the authorization request.
func (s *Service) HandleAuthorizationRequest(ctx context.Context, msg kafka.Message) error {
	tr := otel.Tracer("authorization-service")
	ctx, span := tr.Start(ctx, "HandleAuthorizationRequest")
	defer span.End()

	var reqEvent events.AuthorizationRequested
	if err := json.Unmarshal(msg.Value, &reqEvent); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback is a no-op if tx is committed

	// --- Business Logic ---
	// In a real application, you would check for credit limits, fraud, etc.
	// For this example, we'll just assume it succeeds.
	log.Printf("Authorizing transaction %s for amount %f", reqEvent.TransactionID, reqEvent.Amount)

	// Save authorization status to the database
	_, err = tx.ExecContext(ctx, "INSERT INTO authorizations (transaction_id, amount, status) VALUES ($1, $2, $3)",
		reqEvent.TransactionID, reqEvent.Amount, "SUCCEEDED")
	if err != nil {
		return err
	}

	// --- Transactional Outbox ---
	// Create the success event
	successEvent := events.AuthorizationSucceeded{
		TransactionID: reqEvent.TransactionID,
	}

	// Add the event to the outbox as part of the same transaction
	if err := outbox.AddToOutbox(tx, "authorization-succeeded", reqEvent.TransactionID, successEvent); err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}