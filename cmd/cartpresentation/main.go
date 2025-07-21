package main

import (
	"log/slog"
	"net/http"
	"os"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

func main() {
	mux := http.NewServeMux()

	var (
		cartStore v1.CartStore
		itemStore v1.ItemStore
	)

	err := router.DefaultRouter.Register(&v1.CartPresentationRouter{ItemStore: itemStore, CartStore: cartStore})
	if err != nil {
		slog.Error("Failed to register cart presentation router", "error", err)
		os.Exit(1)
	}

	err = router.DefaultRouter.Build(mux)
	if err != nil {
		slog.Error("Failed to build router", "error", err)
		os.Exit(1)
	}
	slog.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	slog.Warn("Server stopped")
}
