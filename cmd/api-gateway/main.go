package main

import (
	"encoding/json"
	"log"
	"net/http"

	"credit-authorization-ledger/internal/idempotency"
	"credit-authorization-ledger/internal/kafka"
	"credit-authorization-ledger/pkg/events"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	// Initialize AWS session for DynamoDB
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// In a real app, you'd configure the endpoint for local vs. cloud
		// For this project, the SDK defaults and environment variables work.
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoClient := dynamodb.New(sess)
	idempotencyStore := idempotency.NewDynamoDBStore(dynamoClient, "idempotency_keys")

	// Initialize Kafka Producer
	// The broker address is typically loaded from config, but hardcoded here for simplicity in main.
	kafkaProducer, err := kafka.NewProducer([]string{"kafka:9092"})
	if err != nil {
		log.Fatalf("could not create kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Create the final handler, wrapping the business logic with idempotency middleware
	finalHandler := idempotency.Middleware(idempotencyStore)(authorizeHandler(kafkaProducer))
	http.Handle("/authorize", finalHandler)

	log.Println("API Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

// authorizeHandler now accepts the Kafka producer to publish messages.
func authorizeHandler(p *kafka.Producer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req events.AuthorizationRequested
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Basic validation
		if req.TransactionID == "" || req.UserID == "" {
			http.Error(w, "TransactionID and UserID are required", http.StatusBadRequest)
			return
		}

		// Marshal the event struct into a JSON payload for Kafka.
		payload, err := json.Marshal(req)
		if err != nil {
			log.Printf("ERROR: Failed to serialize request for Kafka: %v", err)
			http.Error(w, "Failed to serialize request", http.StatusInternalServerError)
			return
		}

		// Publish the event to the initial topic that the SAGA orchestrator listens to.
		// The transaction ID is used as the Kafka message key to ensure related events
		// are processed in order by the same partition if needed.
		err = p.Publish(r.Context(), "credit-authorization-requested", req.TransactionID, payload)
		if err != nil {
			log.Printf("ERROR: Failed to publish authorization request to Kafka: %v", err)
			http.Error(w, "Failed to publish authorization request", http.StatusInternalServerError)
			return
		}

		log.Printf("Accepted authorization request for transaction: %s", req.TransactionID)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Authorization request accepted"))
	})
}
