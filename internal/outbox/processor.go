package outbox

import (
	"context"
	"database/sql"
	"log"

	"credit-authorization-ledger/internal/kafka"

	"go.opentelemetry.io/otel"
)

type Processor struct {
	db       *sql.DB
	producer *kafka.Producer
}

func NewProcessor(db *sql.DB, producer *kafka.Producer) *Processor {
	return &Processor{db: db, producer: producer}
}

// ProcessOutboxMessages fetches and publishes messages from the outbox table.
func (p *Processor) ProcessOutboxMessages() error {
	tr := otel.Tracer("outbox-processor")
	ctx, span := tr.Start(context.Background(), "ProcessOutboxMessages")
	defer span.End()

	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Select and lock rows to prevent other processor instances from picking them up.
	rows, err := tx.QueryContext(ctx, `
		SELECT id, topic, key, payload FROM outbox ORDER BY created_at ASC LIMIT 10 FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var messages []OutboxMessage
	var idsToDelete []int64

	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.Topic, &msg.Key, &msg.Payload); err != nil {
			log.Printf("error scanning outbox message: %v", err)
			continue
		}
		messages = append(messages, msg)
		idsToDelete = append(idsToDelete, msg.ID)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	if len(messages) == 0 {
		return nil // Nothing to process
	}

	log.Printf("Processing %d messages from outbox", len(messages))

	for _, msg := range messages {
		if err := p.producer.Publish(ctx, msg.Topic, msg.Key, msg.Payload); err != nil {
			// If we can't publish, we rollback and will retry later.
			log.Printf("failed to publish message %d to kafka: %v", msg.ID, err)
			return err
		}
	}

	// All messages were published successfully, now delete them from the outbox.
	_, err = tx.ExecContext(ctx, "DELETE FROM outbox WHERE id = ANY($1)", pq.Array(idsToDelete))
	if err != nil {
		return err
	}

	return tx.Commit()
}