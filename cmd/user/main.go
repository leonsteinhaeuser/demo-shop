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

func main() {
	mux := http.NewServeMux()

	var (
		userStore v1.UserStore = inmem.NewUserInMemStorage()
	)

	err := router.DefaultRouter.Register(v1.NewUserRouter(userStore))
	if err != nil {
		slog.Error("Failed to register user router", "error", err)
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
