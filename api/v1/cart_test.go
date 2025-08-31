package v1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockCartStore implements CartStore interface for testing
type MockCartStore struct {
	carts  map[uuid.UUID]*Cart
	fail   bool
	failOn string
}

func NewMockCartStore() *MockCartStore {
	return &MockCartStore{
		carts: make(map[uuid.UUID]*Cart),
	}
}

func (m *MockCartStore) SetFailure(failOn string) {
	m.fail = true
	m.failOn = failOn
}

func (m *MockCartStore) Create(ctx context.Context, cart *Cart) error {
	if m.fail && m.failOn == "create" {
		return errors.New("mock create error")
	}
	m.carts[cart.ID] = cart
	return nil
}

func (m *MockCartStore) Get(ctx context.Context, id uuid.UUID) (*Cart, error) {
	if m.fail && m.failOn == "get" {
		return nil, errors.New("mock get error")
	}
	cart, exists := m.carts[id]
	if !exists {
		return nil, nil
	}
	return cart, nil
}

func (m *MockCartStore) Update(ctx context.Context, cart *Cart) error {
	if m.fail && m.failOn == "update" {
		return errors.New("mock update error")
	}
	m.carts[cart.ID] = cart
	return nil
}

func (m *MockCartStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.fail && m.failOn == "delete" {
		return errors.New("mock delete error")
	}
	delete(m.carts, id)
	return nil
}

func TestNewCartRouter(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	if router == nil {
		t.Fatal("Expected router to be created")
	}

	if router.Store != store {
		t.Error("Expected store to be set correctly")
	}

	// Check if prometheus counters are initialized
	if router.processedCreateRequests == nil {
		t.Error("Expected processedCreateRequests counter to be initialized")
	}
	if router.processedCreateFailures == nil {
		t.Error("Expected processedCreateFailures counter to be initialized")
	}
}

func TestCartRouter_GetApiVersion(t *testing.T) {
	router := NewCartRouter(NewMockCartStore())
	if router.GetApiVersion() != "v1" {
		t.Errorf("Expected API version v1, got %s", router.GetApiVersion())
	}
}

func TestCartRouter_GetGroup(t *testing.T) {
	router := NewCartRouter(NewMockCartStore())
	if router.GetGroup() != group {
		t.Errorf("Expected group core, got %s", router.GetGroup())
	}
}

func TestCartRouter_GetKind(t *testing.T) {
	router := NewCartRouter(NewMockCartStore())
	if router.GetKind() != "carts" {
		t.Errorf("Expected kind carts, got %s", router.GetKind())
	}
}

func TestCartRouter_Routes(t *testing.T) {
	router := NewCartRouter(NewMockCartStore())
	routes := router.Routes()

	if len(routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(routes))
	}

	// Check if routes contain expected methods
	methods := make(map[string]bool)
	for _, route := range routes {
		methods[route.Method] = true
	}

	expectedMethods := []string{"POST", "GET", "PUT", "DELETE"}
	for _, method := range expectedMethods {
		if !methods[method] {
			t.Errorf("Expected method %s not found in routes", method)
		}
	}
}

func TestCartRouter_createCart_Success(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cart := &Cart{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Items:   []CartItem{},
	}

	req := httptest.NewRequest("POST", "/api/v1/core/carts", nil)
	err := router.createCart(context.Background(), req, cart)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify cart was stored
	storedCart, exists := store.carts[cart.ID]
	if !exists {
		t.Error("Expected cart to be stored")
	}

	if storedCart.OwnerID != cart.OwnerID {
		t.Errorf("Expected OwnerID %s, got %s", cart.OwnerID, storedCart.OwnerID)
	}
}

func TestCartRouter_createCart_NilStore(t *testing.T) {
	router := NewCartRouter(nil)
	cart := &Cart{ID: uuid.New()}

	req := httptest.NewRequest("POST", "/api/v1/core/carts", nil)
	err := router.createCart(context.Background(), req, cart)

	if err == nil {
		t.Error("Expected error for nil store")
	}
}

func TestCartRouter_createCart_EmptyOwnerID(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cart := &Cart{
		ID:      uuid.New(),
		OwnerID: uuid.Nil, // Empty OwnerID is allowed in the current implementation
		Items:   []CartItem{},
	}

	req := httptest.NewRequest("POST", "/api/v1/core/carts", nil)
	err := router.createCart(context.Background(), req, cart)

	if err != nil {
		t.Errorf("Expected no error for empty OwnerID, got %v", err)
	}
}

func TestCartRouter_createCart_StoreError(t *testing.T) {
	store := NewMockCartStore()
	store.SetFailure("create")
	router := NewCartRouter(store)

	cart := &Cart{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Items:   []CartItem{},
	}

	req := httptest.NewRequest("POST", "/api/v1/core/carts", nil)
	err := router.createCart(context.Background(), req, cart)

	if err == nil {
		t.Error("Expected error from store")
	}
}

func TestCartRouter_getCart_Success(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cartID := uuid.New()
	expectedCart := &Cart{
		ID:        cartID,
		OwnerID:   uuid.New(),
		Items:     []CartItem{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.carts[cartID] = expectedCart

	req := httptest.NewRequest("GET", "/api/v1/core/carts/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	cart, err := router.getCart(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cart == nil {
		t.Fatal("Expected cart to be returned")
	}

	if cart.ID != cartID {
		t.Errorf("Expected cart ID %s, got %s", cartID, cart.ID)
	}
}

func TestCartRouter_getCart_NotFound(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cartID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/core/carts/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	cart, err := router.getCart(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error for cart not found, got %v", err)
	}

	if cart != nil {
		t.Error("Expected nil cart for not found")
	}
}

func TestCartRouter_updateCart_Success(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cartID := uuid.New()
	originalCart := &Cart{
		ID:        cartID,
		OwnerID:   uuid.New(),
		Items:     []CartItem{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.carts[cartID] = originalCart

	updatedCart := &Cart{
		ID:      cartID,
		OwnerID: originalCart.OwnerID,
		Items: []CartItem{
			{ItemID: uuid.New(), Quantity: 2},
		},
	}

	req := httptest.NewRequest("PUT", "/api/v1/core/carts/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	err := router.updateCart(context.Background(), req, updatedCart)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify cart was updated
	storedCart := store.carts[cartID]
	if len(storedCart.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(storedCart.Items))
	}
}

func TestCartRouter_deleteCart_Success(t *testing.T) {
	store := NewMockCartStore()
	router := NewCartRouter(store)

	cartID := uuid.New()
	cart := &Cart{
		ID:      cartID,
		OwnerID: uuid.New(),
		Items:   []CartItem{},
	}
	store.carts[cartID] = cart

	req := httptest.NewRequest("DELETE", "/api/v1/core/carts/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	err := router.deleteCart(context.Background(), req, cart)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify cart was deleted
	_, exists := store.carts[cartID]
	if exists {
		t.Error("Expected cart to be deleted")
	}
}
