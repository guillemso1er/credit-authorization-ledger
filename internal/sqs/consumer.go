package sqs

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// MessageHandler is a function that processes a single SQS message.
// It should return an error if the message cannot be processed and needs to be retried.
type MessageHandler func(ctx context.Context, msg *sqs.Message) error

// Consumer struct holds the SQS client and configuration for polling.
type Consumer struct {
	client         *sqs.SQS
	queueURL       string
	handler        MessageHandler
	maxMessages    int64 // Number of messages to fetch at once (1-10)
	pollWaitTime   int64 // Long polling duration (0-20 seconds)
	workers        int   // Number of concurrent workers to process messages
	shutdownSignal chan struct{}
	wg             sync.WaitGroup
}

// NewConsumer creates a new SQS consumer.
func NewConsumer(sess *session.Session, queueURL string, handler MessageHandler) *Consumer {
	return &Consumer{
		client:         sqs.New(sess),
		queueURL:       queueURL,
		handler:        handler,
		maxMessages:    10,               // Default: fetch max messages allowed
		pollWaitTime:   20,               // Default: enable max long polling
		workers:        5,                // Default: 5 concurrent workers
		shutdownSignal: make(chan struct{}),
	}
}

// StartPolling begins polling the SQS queue for messages in a loop.
// It's a blocking call that will run until Shutdown() is called.
func (c *Consumer) StartPolling() {
	log.Printf("Starting SQS polling on queue: %s with %d workers", c.queueURL, c.workers)
	
	// Create a channel to pass received messages to workers
	messages := make(chan *sqs.Message, c.workers)

	// Start worker goroutines
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(i+1, messages)
	}

	// Start the poller loop
	c.poller(messages)

	// Wait for all workers to finish after shutdown is initiated
	c.wg.Wait()
	log.Println("SQS Consumer has shut down.")
}

// poller is the main loop that receives messages from SQS.
func (c *Consumer) poller(messages chan<- *sqs.Message) {
	for {
		select {
		case <-c.shutdownSignal:
			log.Println("Poller received shutdown signal. Stopping message fetching.")
			close(messages) // Close the channel to signal workers to stop
			return
		default:
			log.Println("Polling for messages...")
			receiveInput := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(c.queueURL),
				MaxNumberOfMessages: aws.Int64(c.maxMessages),
				WaitTimeSeconds:     aws.Int64(c.pollWaitTime), // Enables long polling
				// We ask for the ReceiveCount to know how many times a message has been tried.
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameApproximateReceiveCount),
				},
			}

			output, err := c.client.ReceiveMessage(receiveInput)
			if err != nil {
				log.Printf("ERROR: Failed to receive messages from SQS: %v", err)
				// Backoff before retrying to avoid spamming AWS API on error
				time.Sleep(5 * time.Second)
				continue
			}

			if len(output.Messages) > 0 {
				log.Printf("Received %d messages", len(output.Messages))
				for _, msg := range output.Messages {
					messages <- msg
				}
			}
		}
	}
}

// worker processes messages received from the poller.
func (c *Consumer) worker(id int, messages <-chan *sqs.Message) {
	defer c.wg.Done()
	log.Printf("Worker %d started", id)
	
	for msg := range messages {
		log.Printf("[Worker %d] Processing message ID: %s", id, *msg.MessageId)
		
		// Create a context for the handler
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Example timeout

		err := c.handler(ctx, msg)
		if err != nil {
			// **IMPORTANT**: If processing fails, we DO NOT delete the message.
			// SQS will make it visible again in the queue after the "Visibility Timeout" expires.
			// If this happens enough times (based on Redrive Policy), SQS will automatically
			// move it to the configured Dead-Letter Queue (DLQ).
			log.Printf("ERROR: [Worker %d] Failed to process message %s: %v. Message will be retried.", id, *msg.MessageId, err)
		} else {
			// Message was processed successfully, so we delete it from the queue.
			log.Printf("[Worker %d] Message %s processed successfully. Deleting.", id, *msg.MessageId)
			
			if err := c.deleteMessage(msg); err != nil {
				log.Printf("ERROR: [Worker %d] Failed to delete message %s: %v", id, *msg.MessageId, err)
			}
		}
		
		cancel() // Clean up context resources
	}
	
	log.Printf("Worker %d finished.", id)
}

// deleteMessage removes a message from the SQS queue.
func (c *Consumer) deleteMessage(msg *sqs.Message) error {
	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}
	_, err := c.client.DeleteMessage(deleteInput)
	return err
}

// Shutdown gracefully stops the consumer.
func (c *Consumer) Shutdown() {
	log.Println("Initiating SQS consumer shutdown...")
	close(c.shutdownSignal) // Signal the poller to stop
}