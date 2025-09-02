# Build stage
FROM golang:1.19-alpine AS builder

# Argument to specify the service to build
ARG service
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Build the specific service binary
RUN go build -o /app/main ./cmd/${service}

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/app

EXPOSE 8080

CMD ["./app"]