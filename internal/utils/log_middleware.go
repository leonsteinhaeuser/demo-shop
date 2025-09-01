package utils

import (
	"log/slog"
	"net/http"
)

// LogMiddleware creates a middleware that adds logging to HTTP handlers
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Incoming request",
			"method", r.Method,
			"url", r.URL.String(),
			"remote_addr", r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}
