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

// ItemClient implements the ItemStore interface by making HTTP requests to the API server
type ItemClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewItemClient creates a new ItemClient with the given base URL
func NewItemClient(baseURL string) *ItemClient {
	return &ItemClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewItemClientWithHTTPClient creates a new ItemClient with a custom HTTP client
func NewItemClientWithHTTPClient(baseURL string, httpClient *http.Client) *ItemClient {
	return &ItemClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Create implements the ItemStore.Create method
func (i *ItemClient) Create(ctx context.Context, item *apiv1.Item) error {
	url := fmt.Sprintf("%s/api/v1/core/items", i.baseURL)

	jsonData, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the item with the response (which includes generated ID, timestamps, etc.)
	var updatedItem apiv1.Item
	if err := json.NewDecoder(resp.Body).Decode(&updatedItem); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original item object
	*item = updatedItem
	return nil
}

// List implements the ItemStore.List method
func (i *ItemClient) List(ctx context.Context, page, limit int) ([]apiv1.Item, error) {
	url := fmt.Sprintf("%s/api/v1/core/items?page=%d&limit=%d", i.baseURL, page, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var items []apiv1.Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return items, nil
}

// Get implements the ItemStore.Get method
func (i *ItemClient) Get(ctx context.Context, id uuid.UUID) (*apiv1.Item, error) {
	url := fmt.Sprintf("%s/api/v1/core/items/%s", i.baseURL, id.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := i.httpClient.Do(req)
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

	var item apiv1.Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &item, nil
}

// Update implements the ItemStore.Update method
func (i *ItemClient) Update(ctx context.Context, item *apiv1.Item) error {
	url := fmt.Sprintf("%s/api/v1/core/items/%s", i.baseURL, item.ID.String())

	jsonData, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the item with the response
	var updatedItem apiv1.Item
	if err := json.NewDecoder(resp.Body).Decode(&updatedItem); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original item object
	*item = updatedItem
	return nil
}

// Delete implements the ItemStore.Delete method
func (i *ItemClient) Delete(ctx context.Context, id uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/core/items/%s", i.baseURL, id.String())

	// Create a minimal item object for the delete request
	deleteItem := apiv1.Item{ID: id}
	jsonData, err := json.Marshal(deleteItem)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Verify that ItemClient implements the ItemStore interface
var _ apiv1.ItemStore = (*ItemClient)(nil)
