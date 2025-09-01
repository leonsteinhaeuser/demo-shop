package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/cmd/gateway/check"
	"github.com/leonsteinhaeuser/demo-shop/internal/env"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/leonsteinhaeuser/demo-shop/internal/utils"
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

	traceConfig = utils.TraceConfigFromEnv()
)

func main() {
	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	// ping upstream services
	check.Check(ctx,
		envUserServiceURL,
		envCartServiceURL,
		envItemServiceURL,
		envCheckoutServiceURL,
		envCartPresentationServiceURL,
	)

	tracer, shutdown, err := utils.NewTracer(ctx, traceConfig)
	if err != nil {
		slog.Error("Failed to create tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			slog.Error("Failed to shutdown tracer", "error", err)
		}
	}()
	utils.DefaultTracer = tracer

	// Print build information
	slog.Info("API Gateway", "version", version, "commit", commit, "date", date)

	// Create multiplexer and register routes
	mux := http.NewServeMux()

	// Initialize gateway
	v1.NewGateway(
		envUserServiceURL,
		envCartServiceURL,
		envItemServiceURL,
		envCheckoutServiceURL,
		envCartPresentationServiceURL,
		envCookieEncryptionKey,
	).RegisterRoutes(mux)

	err = router.DefaultRouter.Build(mux)
	if err != nil {
		slog.Error("Failed to build router", "error", err)
		os.Exit(1)
	}

	// Configure server with timeouts
	server := &http.Server{
		Addr:           ":8080",
		Handler:        utils.LogMiddleware(router.EnableCorsHeader(utils.TracingMiddleware("gateway")(mux))),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	router.DefaultRouter.SetLiveness(true)
	router.DefaultRouter.SetReady(true)

	utils.StopSignalHandler(
		func(ctx context.Context) {
			slog.Info("API Gateway listening on :8080")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Failed to start server", "error", err)
				ctx.Done()
			}
		},
		func(ctx context.Context) {
			slog.Info("API Gateway shutting down...")
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("Server forced to shutdown", "error", err)
			}
		},
	)
	slog.Warn("Server stopped")
}
