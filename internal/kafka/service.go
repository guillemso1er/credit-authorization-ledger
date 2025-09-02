package ledger

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

func (s *Service) HandleCreditLedger(ctx context.Context, msg kafka.Message) error {
	tr := otel.Tracer("ledger-service")
	ctx, span := tr.Start(ctx, "HandleCreditLedger")
	defer span.End()

	var event events.AuthorizationSucceeded
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Printf("Recording ledger entry for transaction %s", event.TransactionID)
	_, err = tx.ExecContext(ctx,
		"INSERT INTO ledger (transaction_id, entry_type) VALUES ($1, $2)",
		event.TransactionID, "CREDIT_AUTHORIZED")
	if err != nil {
		// Here you would add a `ledger-update-failed` event to the outbox
		return err
	}

	// Add ledger update success event to outbox
	ledgerEvent := events.LedgerUpdateSucceeded{TransactionID: event.TransactionID}
	if err := outbox.AddToOutbox(tx, "ledger-update-succeeded", event.TransactionID, ledgerEvent); err != nil {
		return err
	}

	return tx.Commit()
}