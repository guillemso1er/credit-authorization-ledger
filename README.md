# Credit Authorization & Ledger (Go Microservices Reference)

This project is a reference implementation of a distributed credit authorization and ledger system using Go microservices. It demonstrates several best practices for building resilient, scalable, and observable distributed systems.

## Key Architectural Patterns & Technologies

- **Microservices Architecture**: The system is broken down into small, independent services.
  - `api-gateway`: Public-facing entry point with idempotency checks.
  - `authorization-service`: Handles the core logic of authorizing a credit request.
  - `ledger-service`: Records the transaction in a ledger upon successful authorization.
  - `outbox-processor`: Reliably publishes events from the database to Kafka.
  - `saga-orchestrator`: Manages the distributed transaction across services.
- **Go (Golang)**: The primary language for all microservices.
- **Transactional Outbox Pattern**: Guarantees "at-least-once" event delivery by atomically committing database changes and outgoing events. Implemented with PostgreSQL.
- **SAGA Orchestration Pattern**: Manages long-lived, distributed transactions to ensure data consistency across services without using 2-phase commits.
- **Idempotency**: The API gateway uses DynamoDB to store idempotency keys, preventing duplicate request processing.
- **Reliable Messaging**:
  - **Apache Kafka**: Used as the primary event bus for choreography between services.
  - **Amazon SQS/SNS**: Can be integrated for message queuing with built-in retry and Dead-Letter Queue (DLQ) capabilities.
- **Distributed Tracing**: Integrated with OpenTelemetry for end-to-end request tracing across all services.
- **Containerization**: All services and infrastructure are containerized using **Docker** and orchestrated with **Docker Compose** for easy local development.

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Running Locally

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/YOUR-USER/credit-authorization-ledger.git
    cd credit-authorization-ledger
    ```

2.  **Build and run the services:**
    ```sh
    docker-compose up --build
    ```
    This command will start all the required infrastructure (PostgreSQL, Kafka, DynamoDB) and build and run all the Go microservices.

3.  **Send a request:**
    Use an API client like `curl` or Postman to send a request to the API gateway.

    ```sh
    curl -X POST http://localhost:8080/authorize \
    -H "Content-Type: application/json" \
    -H "Idempotency-Key: $(uuidgen)" \
    -d '{
      "transaction_id": "tx-12345",
      "user_id": "user-6789",
      "amount": 99.99
    }'
    ```

You can then observe the logs from each service to see the SAGA orchestration in action.
