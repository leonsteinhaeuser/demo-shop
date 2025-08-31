package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/leonsteinhaeuser/demo-shop/internal/storage/inmem"
	"github.com/leonsteinhaeuser/demo-shop/internal/utils"
)

// build information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	traceConfig = utils.TraceConfigFromEnv()
)

func main() {
	ctx, cf := context.WithCancel(context.Background())
	defer cf()

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

	slog.Info("Cart Service", "version", version, "commit", commit, "date", date)

	mux := http.NewServeMux()

	var (
		cartStore v1.CartStore = inmem.NewCartInMemStorage()
	)

	err = router.DefaultRouter.Register(v1.NewCartRouter(cartStore))
	if err != nil {
		slog.Error("Failed to register cart router", "error", err)
		os.Exit(1)
	}

	err = router.DefaultRouter.Build(mux)
	if err != nil {
		slog.Error("Failed to build router", "error", err)
		os.Exit(1)
	}
	slog.Info("Starting server on :8080")

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router.EnableCorsHeader(utils.TracingMiddleware("cart")(mux)),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

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
