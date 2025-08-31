package v1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockCheckoutStore implements CheckoutStore interface for testing
type MockCheckoutStore struct {
	checkouts   map[uuid.UUID]*Checkout
	shouldError bool
}

func NewMockCheckoutStore() *MockCheckoutStore {
	return &MockCheckoutStore{
		checkouts: make(map[uuid.UUID]*Checkout),
	}
}

func (m *MockCheckoutStore) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockCheckoutStore) Create(ctx context.Context, checkout *Checkout) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	checkout.CreatedAt = time.Now()
	checkout.UpdatedAt = time.Now()
	m.checkouts[checkout.ID] = checkout
	return nil
}

func (m *MockCheckoutStore) Get(ctx context.Context, id uuid.UUID) (*Checkout, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	checkout, exists := m.checkouts[id]
	if !exists {
		return nil, errors.New("checkout not found")
	}
	return checkout, nil
}

func (m *MockCheckoutStore) Update(ctx context.Context, checkout *Checkout) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	checkout.UpdatedAt = time.Now()
	m.checkouts[checkout.ID] = checkout
	return nil
}

func (m *MockCheckoutStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	delete(m.checkouts, id)
	return nil
}

func TestCheckoutRouter_createCheckout_Success(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkout := &Checkout{
		ID:     uuid.New(),
		CartID: uuid.New(),
		UserID: uuid.New(),
		Total:  99.99,
		Status: "pending",
	}

	req := httptest.NewRequest("POST", "/api/v1/core/checkouts", nil)
	err := router.createCheckout(context.Background(), req, checkout)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify checkout was stored
	storedCheckout, exists := store.checkouts[checkout.ID]
	if !exists {
		t.Error("Expected checkout to be stored")
	}

	if storedCheckout.Status != "pending" {
		t.Errorf("Expected status pending, got %s", storedCheckout.Status)
	}
}

func TestCheckoutRouter_createCheckout_StoreError(t *testing.T) {
	store := NewMockCheckoutStore()
	store.SetError(true)
	router := NewCheckoutRouter(store)

	checkout := &Checkout{
		ID:     uuid.New(),
		CartID: uuid.New(),
		UserID: uuid.New(),
		Total:  99.99,
		Status: "pending",
	}

	req := httptest.NewRequest("POST", "/api/v1/core/checkouts", nil)
	err := router.createCheckout(context.Background(), req, checkout)

	if err == nil {
		t.Error("Expected error from store")
	}
}

func TestCheckoutRouter_createCheckout_InvalidUserID(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkout := &Checkout{
		ID:     uuid.New(),
		CartID: uuid.New(),
		UserID: uuid.Nil, // Empty UserID should cause error
		Total:  99.99,
		Status: "pending",
	}

	req := httptest.NewRequest("POST", "/api/v1/core/checkouts", nil)
	err := router.createCheckout(context.Background(), req, checkout)

	if err == nil {
		t.Error("Expected error for empty UserID")
	}
}

func TestCheckoutRouter_createCheckout_InvalidCartID(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkout := &Checkout{
		ID:     uuid.New(),
		CartID: uuid.Nil, // Empty CartID should cause error
		UserID: uuid.New(),
		Total:  99.99,
		Status: "pending",
	}

	req := httptest.NewRequest("POST", "/api/v1/core/checkouts", nil)
	err := router.createCheckout(context.Background(), req, checkout)

	if err == nil {
		t.Error("Expected error for empty CartID")
	}
}

func TestCheckoutRouter_getCheckout_Success(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkoutID := uuid.New()
	checkout := &Checkout{
		ID:        checkoutID,
		CartID:    uuid.New(),
		UserID:    uuid.New(),
		Total:     99.99,
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.checkouts[checkoutID] = checkout

	req := httptest.NewRequest("GET", "/api/v1/core/checkouts/"+checkoutID.String(), nil)
	req.SetPathValue("id", checkoutID.String())

	retrievedCheckout, err := router.getCheckout(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedCheckout == nil {
		t.Fatal("Expected checkout to be returned")
	}

	if retrievedCheckout.ID != checkoutID {
		t.Errorf("Expected checkout ID %s, got %s", checkoutID, retrievedCheckout.ID)
	}

	if retrievedCheckout.Status != "completed" {
		t.Errorf("Expected status completed, got %s", retrievedCheckout.Status)
	}
}

func TestCheckoutRouter_getCheckout_NotFound(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkoutID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/core/checkouts/"+checkoutID.String(), nil)
	req.SetPathValue("id", checkoutID.String())

	checkout, err := router.getCheckout(context.Background(), req)

	if err == nil {
		t.Error("Expected error for non-existent checkout")
	}

	if checkout != nil {
		t.Error("Expected nil checkout for non-existent ID")
	}
}

func TestCheckoutRouter_updateCheckout_Success(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkoutID := uuid.New()
	originalCheckout := &Checkout{
		ID:        checkoutID,
		CartID:    uuid.New(),
		UserID:    uuid.New(),
		Total:     99.99,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.checkouts[checkoutID] = originalCheckout

	updatedCheckout := &Checkout{
		ID:     checkoutID,
		CartID: originalCheckout.CartID,
		UserID: originalCheckout.UserID,
		Total:  129.99,
		Status: "completed",
	}

	req := httptest.NewRequest("PUT", "/api/v1/core/checkouts/"+checkoutID.String(), nil)
	req.SetPathValue("id", checkoutID.String())

	err := router.updateCheckout(context.Background(), req, updatedCheckout)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify checkout was updated
	storedCheckout := store.checkouts[checkoutID]
	if storedCheckout.Status != "completed" {
		t.Errorf("Expected updated status completed, got %s", storedCheckout.Status)
	}

	if storedCheckout.Total != 129.99 {
		t.Errorf("Expected updated total 129.99, got %f", storedCheckout.Total)
	}
}

func TestCheckoutRouter_deleteCheckout_Success(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	checkoutID := uuid.New()
	checkout := &Checkout{
		ID:     checkoutID,
		CartID: uuid.New(),
		UserID: uuid.New(),
		Total:  99.99,
		Status: "pending",
	}
	store.checkouts[checkoutID] = checkout

	req := httptest.NewRequest("DELETE", "/api/v1/core/checkouts/"+checkoutID.String(), nil)
	req.SetPathValue("id", checkoutID.String())

	err := router.deleteCheckout(context.Background(), req, checkout)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify checkout was deleted
	_, exists := store.checkouts[checkoutID]
	if exists {
		t.Error("Expected checkout to be deleted")
	}
}

func TestCheckoutRouter_deleteCheckout_NilCheckout(t *testing.T) {
	store := NewMockCheckoutStore()
	router := NewCheckoutRouter(store)

	req := httptest.NewRequest("DELETE", "/api/v1/core/checkouts/"+uuid.New().String(), nil)

	err := router.deleteCheckout(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error for nil checkout")
	}
}
