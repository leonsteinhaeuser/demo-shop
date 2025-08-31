package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/env"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

// Build information set by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	envUserServiceURL             = env.StringEnvOrDefault("USER_SERVICE_URL", "http://localhost:8084")
	envCartServiceURL             = env.StringEnvOrDefault("CART_SERVICE_URL", "http://localhost:8082")
	envItemServiceURL             = env.StringEnvOrDefault("ITEM_SERVICE_URL", "http://localhost:8081")
	envCheckoutServiceURL         = env.StringEnvOrDefault("CHECKOUT_SERVICE_URL", "http://localhost:8085")
	envCartPresentationServiceURL = env.StringEnvOrDefault("CART_PRESENTATION_SERVICE_URL", "http://localhost:8083")
	envCookieEncryptionKey        = env.BytesEnvOrDefault("COOKIE_ENCRYPTION_KEY", []byte("a_random_secret_key"))
)

func main() {
	// Print build information
	slog.Info("API Gateway", "version", version, "commit", commit, "date", date)

	// Initialize gateway
	gateway := v1.NewGateway(envUserServiceURL, envCartServiceURL, envItemServiceURL, envCheckoutServiceURL, envCartPresentationServiceURL, envCookieEncryptionKey)

	// Create multiplexer and register routes
	mux := http.NewServeMux()
	gateway.RegisterRoutes(mux)

	// Configure server with timeouts
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router.EnableCorsHeader(mux),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Println("API Gateway starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
