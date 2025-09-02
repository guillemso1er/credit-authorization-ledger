package main

import (
	"log"
	"net/http"
	"credit-authorization-ledger/internal/idempotency"
	"credit-authorization-ledger/internal/server"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoClient := dynamodb.New(sess)
	idempotencyStore := idempotency.NewDynamoDBStore(dynamoClient, "idempotency_keys")

	http.Handle("/authorize", idempotency.Middleware(idempotencyStore)(authorizeHandler()))

	log.Println("API Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

func authorizeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would publish an event to Kafka
		// to start the SAGA orchestration.
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Authorization request accepted"))
	})
}
