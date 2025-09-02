package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		// Topics will be specified in the Consume method
	})
	return &Consumer{reader: reader}
}

func (c *Consumer) Consume(topics []string, handler MessageHandler) {
	c.reader.Config().Topic = topics[0] // Simplified for single topic, adjust for multi-topic
	if len(topics) > 1 {
		// The library supports this, but our simple loop needs adjustment.
		// For a real app, you might run a goroutine per topic or use a different consumption pattern.
		log.Println("Warning: This consumer is simplified for a single topic from the list.")
	}

	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}
		log.Printf("message received: Topic=%s, Key=%s", msg.Topic, string(msg.Key))
		if err := handler(context.Background(), msg); err != nil {
			log.Printf("error handling message: %v", err)
			// Implement retry/DLQ logic here if needed at the consumer level
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}