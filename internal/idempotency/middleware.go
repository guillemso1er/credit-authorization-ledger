package idempotency

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the response body and status code.
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
		statusCode:     http.StatusOK, // Default status code
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type KeyStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

func Middleware(store KeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idempotencyKey := r.Header.Get("Idempotency-Key")
			if idempotencyKey == "" {
				// If no key, proceed without idempotency checks.
				next.ServeHTTP(w, r)
				return
			}

			// Check if the response is already cached.
			cachedResponse, err := store.Get(r.Context(), idempotencyKey)
			if err == nil && cachedResponse != "" {
				// Key found, return the cached response.
				w.WriteHeader(http.StatusAccepted) // Or whatever the original status was
				w.Write([]byte(cachedResponse))
				return
			}

			// Key not found, proceed with the request and capture the response.
			wrappedWriter := newResponseWriter(w)
			next.ServeHTTP(wrappedWriter, r)

			// Store the actual response body in the idempotency store.
			responseBody := wrappedWriter.body.String()
			if responseBody != "" {
				// Store the response for 24 hours.
				// The context from the original request might be done, so use a background context for storage.
				store.Set(context.Background(), idempotencyKey, responseBody, 24*time.Hour)
			}
		})
	}
}