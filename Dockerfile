# Build stage
FROM golang:1.19-alpine AS builder

# Argument to specify the service to build
ARG service
WORKDIR /app

# Copy module files and download dependencies first to leverage Docker cache
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the specific service binary
# CGO_ENABLED=0 is important for creating a static binary that works in a minimal alpine image
# -ldflags="-w -s" strips debug information to create a smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/${service}

# Final stage
FROM alpine:latest

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/app

# Set ownership to the non-root user
RUN chown appuser:appgroup /app/app

# Switch to the non-root user
USER appuser

EXPOSE 8080

CMD ["./app"]