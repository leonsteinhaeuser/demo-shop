package main

import (
	"log/slog"
	"net/http"
	"os"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/leonsteinhaeuser/demo-shop/internal/storage/inmem"
)

func main() {
	mux := http.NewServeMux()

	var (
		itemStore v1.ItemStore = inmem.NewItemInMemStorage()
	)

	err := router.DefaultRouter.Register(v1.NewItemRouter(itemStore))
	if err != nil {
		slog.Error("Failed to register item router", "error", err)
		os.Exit(1)
	}

	err = router.DefaultRouter.Build(mux)
	if err != nil {
		slog.Error("Failed to build router", "error", err)
		os.Exit(1)
	}
	slog.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router.EnableCorsHeader(mux)); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	slog.Warn("Server stopped")
}
