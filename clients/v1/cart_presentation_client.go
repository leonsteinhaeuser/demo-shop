package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

// CartPresentationClient provides access to cart presentation endpoints
type CartPresentationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCartPresentationClient creates a new CartPresentationClient with the given base URL
func NewCartPresentationClient(baseURL string) *CartPresentationClient {
	return &CartPresentationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewCartPresentationClientWithHTTPClient creates a new CartPresentationClient with a custom HTTP client
func NewCartPresentationClientWithHTTPClient(baseURL string, httpClient *http.Client) *CartPresentationClient {
	return &CartPresentationClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// GetCartPresentation retrieves the full cart presentation with item details and pricing
func (c *CartPresentationClient) GetCartPresentation(ctx context.Context, cartID uuid.UUID) (*apiv1.CartPresentation, error) {
	url := fmt.Sprintf("%s/api/v1/presentation/cart/%s", c.baseURL, cartID.String())

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

	var cartPresentation apiv1.CartPresentation
	if err := json.NewDecoder(resp.Body).Decode(&cartPresentation); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cartPresentation, nil
}
