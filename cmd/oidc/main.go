package main

import (
	"log"
	"net/http"

	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

func main() {
	// Create OIDC configuration
	config := &v1.OIDCConfig{
		Issuer:        "http://localhost:8080",
		Port:          8080,
		AllowInsecure: true,
	}

	// Create OIDC router
	oidcRouter, err := v1.NewOIDCRouter(config)
	if err != nil {
		log.Fatalf("Failed to create OIDC router: %v", err)
	}

	// Register the OIDC router
	if err := router.DefaultRouter.Register(oidcRouter); err != nil {
		log.Fatalf("Failed to register OIDC router: %v", err)
	}

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Build the router
	if err := router.DefaultRouter.Build(mux); err != nil {
		log.Fatalf("Failed to build router: %v", err)
	}

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add root redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/api/metadata", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, r)
	})

	log.Printf("Starting OIDC server on :%d", config.Port)
	log.Printf("Issuer: %s", config.Issuer)
	log.Printf("Discovery endpoint: %s/.well-known/openid_configuration", config.Issuer)
	log.Printf("API metadata: %s/api/metadata", config.Issuer)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
