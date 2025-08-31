package v1

import (
	"bytes"
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
	ctx, span := utils.SpanFromContext(ctx, "cart.client.create")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/carts", c.baseURL)

	jsonData, err := json.Marshal(cart)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		span.RecordError(err)
		return err
	}

	// Update the cart with the response (which includes generated ID, timestamps, etc.)
	var updatedCart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&updatedCart); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original cart object
	*cart = updatedCart
	span.SetAttributes(attribute.String("cart.id", cart.ID.String()))
	return nil
}

// Get implements the CartStore.Get method
func (c *CartClient) Get(ctx context.Context, id uuid.UUID) (*apiv1.Cart, error) {
	ctx, span := utils.SpanFromContext(ctx, "cart.client.get")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, id.String())

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

	var cart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cart, nil
}

// Update implements the CartStore.Update method
func (c *CartClient) Update(ctx context.Context, cart *apiv1.Cart) error {
	ctx, span := utils.SpanFromContext(ctx, "cart.client.update")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, cart.ID.String())

	jsonData, err := json.Marshal(cart)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		span.RecordError(err)
		return err
	}

	// Update the cart with the response
	var updatedCart apiv1.Cart
	if err := json.NewDecoder(resp.Body).Decode(&updatedCart); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original cart object
	*cart = updatedCart
	return nil
}

// Delete implements the CartStore.Delete method
func (c *CartClient) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := utils.SpanFromContext(ctx, "cart.client.delete")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/carts/%s", c.baseURL, id.String())

	// Create a minimal cart object for the delete request
	deleteCart := apiv1.Cart{ID: id}
	jsonData, err := json.Marshal(deleteCart)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		span.RecordError(err)
		return err
	}

	return nil
}

// Verify that CartClient implements the CartStore interface
var _ apiv1.CartStore = (*CartClient)(nil)
