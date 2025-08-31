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
)

// UserClient implements the UserStore interface by making HTTP requests to the API server
type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewUserClient creates a new UserClient with the given base URL
func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewUserClientWithHTTPClient creates a new UserClient with a custom HTTP client
func NewUserClientWithHTTPClient(baseURL string, httpClient *http.Client) *UserClient {
	return &UserClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Create implements the UserStore.Create method
func (u *UserClient) Create(ctx context.Context, user *apiv1.UserModificationRequest) error {
	ctx, span := utils.SpanFromContext(ctx, "user.client.create")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/users", u.baseURL)

	jsonData, err := json.Marshal(user)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the user with the response (which includes generated ID, timestamps, etc.)
	var updatedUser apiv1.UserModificationRequest
	if err := json.NewDecoder(resp.Body).Decode(&updatedUser); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original user object
	*user = updatedUser
	return nil
}

// List implements the UserStore.List method
func (u *UserClient) List(ctx context.Context, page, limit int) ([]apiv1.User, error) {
	ctx, span := utils.SpanFromContext(ctx, "user.client.list")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/users?page=%d&limit=%d", u.baseURL, page, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var users []apiv1.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// Get implements the UserStore.Get method
func (u *UserClient) Get(ctx context.Context, id uuid.UUID) (*apiv1.User, error) {
	ctx, span := utils.SpanFromContext(ctx, "user.client.get")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/users/%s", u.baseURL, id.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		span.RecordError(fmt.Errorf("user not found: %s", id.String()))
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var user apiv1.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// Update implements the UserStore.Update method
func (u *UserClient) Update(ctx context.Context, user *apiv1.UserModificationRequest) error {
	ctx, span := utils.SpanFromContext(ctx, "user.client.update")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/users/%s", u.baseURL, user.ID.String())

	jsonData, err := json.Marshal(user)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update the user with the response
	var updatedUser apiv1.UserModificationRequest
	if err := json.NewDecoder(resp.Body).Decode(&updatedUser); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the original user object
	*user = updatedUser
	return nil
}

// Delete implements the UserStore.Delete method
func (u *UserClient) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := utils.SpanFromContext(ctx, "user.client.delete")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/core/users/%s", u.baseURL, id.String())

	// Create a minimal delete request object
	deleteReq := apiv1.UserDeleteRequest{ID: id}
	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Verify that UserClient implements the UserStore interface
var _ apiv1.UserStore = (*UserClient)(nil)
