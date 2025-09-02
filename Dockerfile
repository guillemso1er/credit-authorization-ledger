# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main ./cmd/...

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/app

EXPOSE 8080

CMD ["./app"]
