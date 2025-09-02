module credit-authorization-ledger

go 1.19

require (
	github.com/aws/aws-sdk-go v1.44.128
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/lib/pq v1.10.7
	github.com/segmentio/kafka-go v0.4.38
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.11.2
	go.opentelemetry.io/otel/sdk v1.11.2
	go.opentelemetry.io/otel/trace v1.11.2
)

// In a real project, a go.sum file would also be generated
// by running `go mod tidy`. It is not included here as it is
// machine-generated and very long.
