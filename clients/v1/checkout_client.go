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

// CheckoutClient implements the CheckoutStore interface by making HTTP requests to the API server
type CheckoutClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCheckoutClient creates a new CheckoutClient with the given base URL
func NewCheckoutClient(baseURL string) *CheckoutClient {
	return &CheckoutClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewCheckoutClientWithHTTPClient creates a new CheckoutClient with a custom HTTP client
func NewCheckoutClientWithHTTPClient(baseURL string, httpClient *http.Client) *CheckoutClient {
	return &CheckoutClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Create implements the CheckoutStore.Create method
func (c *CheckoutClient) Create(ctx context.Context, checkout *apiv1.Checkout) error {
	url := fmt.Sprintf("%s/api/v1/core/checkouts", c.baseURL)

	jsonData, err := json.Marshal(checkout)
	if err != nil {
		return fmt.Errorf("failed to marshal checkout: %w", err)
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

	// Update the checkout with the response (which includes generated ID, timestamps, etc.)
	var updatedCheckout apiv1.Checkout
	if err := json.NewDecoder(resp.Body).Decode(&updatedCheckout); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original checkout object
	*checkout = updatedCheckout
	return nil
}

// Get implements the CheckoutStore.Get method
func (c *CheckoutClient) Get(ctx context.Context, id uuid.UUID) (*apiv1.Checkout, error) {
	url := fmt.Sprintf("%s/api/v1/core/checkouts/%s", c.baseURL, id.String())

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

	var checkout apiv1.Checkout
	if err := json.NewDecoder(resp.Body).Decode(&checkout); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &checkout, nil
}

// Update implements the CheckoutStore.Update method
func (c *CheckoutClient) Update(ctx context.Context, checkout *apiv1.Checkout) error {
	url := fmt.Sprintf("%s/api/v1/core/checkouts/%s", c.baseURL, checkout.ID.String())

	jsonData, err := json.Marshal(checkout)
	if err != nil {
		return fmt.Errorf("failed to marshal checkout: %w", err)
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

	// Update the checkout with the response
	var updatedCheckout apiv1.Checkout
	if err := json.NewDecoder(resp.Body).Decode(&updatedCheckout); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original checkout object
	*checkout = updatedCheckout
	return nil
}

// Delete implements the CheckoutStore.Delete method
func (c *CheckoutClient) Delete(ctx context.Context, id uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/core/checkouts/%s", c.baseURL, id.String())

	// Create a minimal checkout object for the delete request
	deleteCheckout := apiv1.Checkout{ID: id}
	jsonData, err := json.Marshal(deleteCheckout)
	if err != nil {
		return fmt.Errorf("failed to marshal checkout: %w", err)
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

// Verify that CheckoutClient implements the CheckoutStore interface
var _ apiv1.CheckoutStore = (*CheckoutClient)(nil)
