package idempotency

import (
	"context"
	"net/http"
	"time"
)

type KeyStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

func Middleware(store KeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idempotencyKey := r.Header.Get("Idempotency-Key")
			if idempotencyKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			response, err := store.Get(r.Context(), idempotencyKey)
			if err == nil && response != "" {
				// Key found, return the cached response
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
				return
			}

			// Key not found, proceed with the request
			next.ServeHTTP(w, r)

			// In a real scenario, you'd capture the response and store it.
			// This is a simplified example.
			store.Set(r.Context(), idempotencyKey, "processed", 24*time.Hour)
		})
	}
}
