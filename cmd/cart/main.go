package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/leonsteinhaeuser/demo-shop/internal/storage/inmem"
)

// build information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	slog.Info("Cart Service", "version", version, "commit", commit, "date", date)

	mux := http.NewServeMux()

	var (
		cartStore v1.CartStore = inmem.NewCartInMemStorage()
	)

	err := router.DefaultRouter.Register(v1.NewCartRouter(cartStore))
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
		Handler:        router.EnableCorsHeader(mux),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	slog.Warn("Server stopped")
}
