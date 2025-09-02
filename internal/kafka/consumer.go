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
	// Topics are now set in the Consume method
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
	})
	return &Consumer{reader: reader}
}

// Consume starts consuming messages from the given topics.
// It takes a context to allow for graceful shutdown.
func (c *Consumer) Consume(ctx context.Context, topics []string, handler MessageHandler) {
	// The kafka-go reader can handle multiple topics if you list them at creation,
	// but a better pattern for distinct logic is separate consumers.
	// For this SAGA orchestrator where one handler manages multiple topics, this is okay.
	// We'll reconfigure the reader for the topics it needs to consume.
	c.reader.Close() // Close previous config if any
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.reader.Config().Brokers,
		GroupID: c.reader.Config().GroupID,
		GroupTopics: topics,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})


	log.Printf("Consumer started for topics: %v", topics)
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, indicating a shutdown.
			log.Println("Shutdown signal received, stopping consumer.")
			return
		default:
			// Use FetchMessage which respects context cancellation.
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				// If context is cancelled, reader will return an error. Check for it.
				if ctx.Err() != nil {
					return
				}
				log.Printf("error fetching message: %v", err)
				continue // Or break, depending on desired behavior for errors
			}

			log.Printf("message received: Topic=%s, Key=%s", msg.Topic, string(msg.Key))

			// Process the message
			if err := handler(ctx, msg); err != nil {
				log.Printf("error handling message: %v", err)
				// In a real app, decide whether to commit or not based on the error.
				// For now, we continue and let the next message be read.
			}

			// Commit the message offset.
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}