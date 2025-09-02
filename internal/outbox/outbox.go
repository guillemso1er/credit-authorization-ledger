package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type OutboxMessage struct {
	ID        int64
	Topic     string
	Key       string
	Payload   []byte
	CreatedAt time.Time
}

func AddToOutbox(tx *sql.Tx, topic, key string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO outbox (topic, key, payload)
		VALUES ($1, $2, $3)
	`, topic, key, payloadBytes)

	return err
}

// In a separate `outbox-processor` service:
func ProcessOutboxMessages(db *sql.DB, kafkaProducer KafkaProducer) {
	rows, err := db.Query("SELECT id, topic, key, payload FROM outbox ORDER BY created_at ASC LIMIT 100")
	if err != nil {
		// Handle error
		return
	}
	defer rows.Close()

	var messages []OutboxMessage
	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.Topic, &msg.Key, &msg.Payload); err != nil {
			// Handle error
			continue
		}
		messages = append(messages, msg)
	}

	for _, msg := range messages {
		// Publish to Kafka
		kafkaProducer.Publish(msg.Topic, msg.Key, msg.Payload)

		// Delete from outbox
		db.Exec("DELETE FROM outbox WHERE id = $1", msg.ID)
	}
}

type KafkaProducer interface {
	Publish(topic, key string, payload []byte) error
}
