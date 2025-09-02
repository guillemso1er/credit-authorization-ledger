package sqs

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Consumer struct {
	client   *sqs.SQS
	queueURL string
}

func NewConsumer(sess *session.Session, queueURL string) *Consumer {
	return &Consumer{
		client:   sqs.New(sess),
		queueURL: queueURL,
	}
}

// StartPolling begins polling the SQS queue for messages.
// This is where you would implement logic with retries and DLQ handling.
func (c *Consumer) StartPolling() {
	log.Printf("Starting to poll SQS queue: %s", c.queueURL)
	// In a real implementation:
	// 1. Loop forever.
	// 2. Receive messages from SQS.
	// 3. For each message, attempt to process it.
	// 4. If successful, delete the message from the queue.
	// 5. If it fails, allow the message visibility timeout to expire so it can be retried.
	// 6. After N retries (based on the ReceiveCount attribute), the message is automatically moved to the configured DLQ.
}