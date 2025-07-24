package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

// CartClient implements the CartStore interface by making HTTP requests to the API server
type CartClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCartClient creates a new CartClient with the given base URL
func NewCartClient(baseURL string) *CartClient {
	return &CartClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewCartClientWithHTTPClient creates a new CartClient with a custom HTTP client
func NewCartClientWithHTTPClient(baseURL string, httpClient *http.Client) *CartClient {
	return &CartClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Create implements the CartStore.Create method
func (c *CartClient) Create(ctx context.Context, cart *apiv1.Cart) error {
	url := fmt.Sprintf("%s/api/v1/core/carts", c.baseURL)

	jsonData, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the cart with the response (which includes generated ID, timestamps, etc.)
	var updatedCart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&updatedCart); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original cart object
	*cart = updatedCart
	return nil
}

// Get implements the CartStore.Get method
func (c *CartClient) Get(ctx context.Context, id uuid.UUID) (*apiv1.Cart, error) {
	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, id.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cart, nil
}

// Update implements the CartStore.Update method
func (c *CartClient) Update(ctx context.Context, cart *apiv1.Cart) error {
	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, cart.ID.String())

	jsonData, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the cart with the response
	var updatedCart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&updatedCart); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original cart object
	*cart = updatedCart
	return nil
}

// Delete implements the CartStore.Delete method
func (c *CartClient) Delete(ctx context.Context, id uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, id.String())

	// Create a minimal cart object for the delete request
	deleteCart := apiv1.Cart{ID: id}
	jsonData, err := json.Marshal(deleteCart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Verify that CartClient implements the CartStore interface
var _ apiv1.CartStore = (*CartClient)(nil)
