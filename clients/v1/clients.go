// Package v1 provides HTTP clients for the demo-shop v1 API
package v1

import (
	"net/http"

	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

// Config holds configuration for all API clients
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

// Clients contains all available API clients
type Clients struct {
	Cart             apiv1.CartStore
	Item             apiv1.ItemStore
	User             apiv1.UserStore
	CartPresentation *CartPresentationClient
}

// NewClients creates a new set of API clients with the given configuration
func NewClients(config Config) *Clients {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Clients{
		Cart:             NewCartClientWithHTTPClient(config.BaseURL, httpClient),
		Item:             NewItemClientWithHTTPClient(config.BaseURL, httpClient),
		User:             NewUserClientWithHTTPClient(config.BaseURL, httpClient),
		CartPresentation: NewCartPresentationClientWithHTTPClient(config.BaseURL, httpClient),
	}
}

// NewDefaultClients creates a new set of API clients with default HTTP client
func NewDefaultClients(baseURL string) *Clients {
	return NewClients(Config{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	})
}
