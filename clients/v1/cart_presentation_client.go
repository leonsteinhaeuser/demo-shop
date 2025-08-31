package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
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
	ctx, span := utils.SpanFromContext(ctx, "cart_presentation.client.get")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/presentation/cart/%s", c.baseURL, cartID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Inject trace context into request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		span.RecordError(err)
		return nil, err
	}

	var cartPresentation apiv1.CartPresentation
	if err := json.NewDecoder(resp.Body).Decode(&cartPresentation); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(
		attribute.Int("cart_presentation.items_count", len(cartPresentation.Items)),
		attribute.Float64("cart_presentation.total_price", cartPresentation.TotalPrice),
	)

	return &cartPresentation, nil
}
