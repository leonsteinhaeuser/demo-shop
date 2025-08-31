package v1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockCartPresentationItemStore implements ItemStore for testing
type MockCartPresentationItemStore struct {
	items  map[uuid.UUID]*Item
	fail   bool
	failOn string
}

func NewMockCartPresentationItemStore() *MockCartPresentationItemStore {
	return &MockCartPresentationItemStore{
		items: make(map[uuid.UUID]*Item),
	}
}

func (m *MockCartPresentationItemStore) SetFailure(failOn string) {
	m.fail = true
	m.failOn = failOn
}

func (m *MockCartPresentationItemStore) Create(ctx context.Context, item *Item) error {
	if m.fail && m.failOn == "item_create" {
		return errors.New("mock item create error")
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockCartPresentationItemStore) List(ctx context.Context, page, limit int) ([]Item, error) {
	if m.fail && m.failOn == "item_list" {
		return nil, errors.New("mock item list error")
	}
	items := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		items = append(items, *item)
	}
	return items, nil
}

func (m *MockCartPresentationItemStore) Get(ctx context.Context, id uuid.UUID) (*Item, error) {
	if m.fail && m.failOn == "item_get" {
		return nil, errors.New("mock item get error")
	}
	item, exists := m.items[id]
	if !exists {
		return nil, nil
	}
	return item, nil
}

func (m *MockCartPresentationItemStore) Update(ctx context.Context, item *Item) error {
	if m.fail && m.failOn == "item_update" {
		return errors.New("mock item update error")
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockCartPresentationItemStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.fail && m.failOn == "item_delete" {
		return errors.New("mock item delete error")
	}
	delete(m.items, id)
	return nil
}

// MockCartPresentationCartStore implements CartStore for testing
type MockCartPresentationCartStore struct {
	carts  map[uuid.UUID]*Cart
	fail   bool
	failOn string
}

func NewMockCartPresentationCartStore() *MockCartPresentationCartStore {
	return &MockCartPresentationCartStore{
		carts: make(map[uuid.UUID]*Cart),
	}
}

func (m *MockCartPresentationCartStore) SetFailure(failOn string) {
	m.fail = true
	m.failOn = failOn
}

func (m *MockCartPresentationCartStore) Create(ctx context.Context, cart *Cart) error {
	if m.fail && m.failOn == "cart_create" {
		return errors.New("mock cart create error")
	}
	m.carts[cart.ID] = cart
	return nil
}

func (m *MockCartPresentationCartStore) Get(ctx context.Context, id uuid.UUID) (*Cart, error) {
	if m.fail && m.failOn == "cart_get" {
		return nil, errors.New("mock cart get error")
	}
	cart, exists := m.carts[id]
	if !exists {
		return nil, nil
	}
	return cart, nil
}

func (m *MockCartPresentationCartStore) Update(ctx context.Context, cart *Cart) error {
	if m.fail && m.failOn == "cart_update" {
		return errors.New("mock cart update error")
	}
	m.carts[cart.ID] = cart
	return nil
}

func (m *MockCartPresentationCartStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.fail && m.failOn == "cart_delete" {
		return errors.New("mock cart delete error")
	}
	delete(m.carts, id)
	return nil
}

func TestNewCartPresentationRouter(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)

	if router == nil {
		t.Fatal("Expected router to be created")
	}

	if router.ItemStore == nil {
		t.Error("Expected ItemStore to be set correctly")
	}

	if router.CartStore == nil {
		t.Error("Expected CartStore to be set correctly")
	}

	// Check if prometheus counters are initialized
	if router.processedGetRequests == nil {
		t.Error("Expected processedGetRequests counter to be initialized")
	}
}

func TestCartPresentationRouter_GetApiVersion(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)
	if router.GetApiVersion() != "v1" {
		t.Errorf("Expected API version v1, got %s", router.GetApiVersion())
	}
}

func TestCartPresentationRouter_GetGroup(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)
	if router.GetGroup() != "presentation" {
		t.Errorf("Expected group presentation, got %s", router.GetGroup())
	}
}

func TestCartPresentationRouter_GetKind(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)
	if router.GetKind() != "cart" {
		t.Errorf("Expected kind cart, got %s", router.GetKind())
	}
}

func TestCartPresentationRouter_getCartPresentation_Success(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)

	// Create test data
	cartID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	// Add items to store
	item1 := &Item{
		ID:          itemID1,
		Name:        "Test Item 1",
		Description: "Test Description 1",
		Price:       10.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	item2 := &Item{
		ID:          itemID2,
		Name:        "Test Item 2",
		Description: "Test Description 2",
		Price:       25.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	itemStore.items[itemID1] = item1
	itemStore.items[itemID2] = item2

	// Add cart to store
	cart := &Cart{
		ID:      cartID,
		OwnerID: uuid.New(),
		Items: []CartItem{
			{ItemID: itemID1, Quantity: 2},
			{ItemID: itemID2, Quantity: 1},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cartStore.carts[cartID] = cart

	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if presentation == nil {
		t.Fatal("Expected cart presentation to be returned")
	}

	if len(presentation.Items) != 2 {
		t.Errorf("Expected 2 items in presentation, got %d", len(presentation.Items))
	}

	expectedTotal := (10.99 * 2) + (25.99 * 1) // 47.97
	if presentation.TotalPrice != expectedTotal {
		t.Errorf("Expected total price %.2f, got %.2f", expectedTotal, presentation.TotalPrice)
	}

	// Check first item
	firstItem := presentation.Items[0]
	if firstItem.Item.ID != itemID1 {
		t.Errorf("Expected first item ID %s, got %s", itemID1, firstItem.Item.ID)
	}
	if firstItem.Quantity != 2 {
		t.Errorf("Expected first item quantity 2, got %d", firstItem.Quantity)
	}
	expectedFirstTotal := 10.99 * 2
	if firstItem.TotalPrice != expectedFirstTotal {
		t.Errorf("Expected first item total %.2f, got %.2f", expectedFirstTotal, firstItem.TotalPrice)
	}
}

func TestCartPresentationRouter_getCartPresentation_EmptyCart(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)

	cartID := uuid.New()
	cart := &Cart{
		ID:        cartID,
		OwnerID:   uuid.New(),
		Items:     []CartItem{}, // Empty cart
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cartStore.carts[cartID] = cart

	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if presentation == nil {
		t.Fatal("Expected cart presentation to be returned")
	}

	if len(presentation.Items) != 0 {
		t.Errorf("Expected 0 items in presentation, got %d", len(presentation.Items))
	}

	if presentation.TotalPrice != 0.0 {
		t.Errorf("Expected total price 0.0, got %.2f", presentation.TotalPrice)
	}
}

func TestCartPresentationRouter_getCartPresentation_CartNotFound(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)

	cartID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err == nil {
		t.Error("Expected error for cart not found")
	}

	if presentation != nil {
		t.Error("Expected nil presentation for cart not found")
	}
}

func TestCartPresentationRouter_getCartPresentation_ItemNotFound(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, cartStore)

	cartID := uuid.New()
	itemID := uuid.New()

	// Add cart but not the item it references
	cart := &Cart{
		ID:      cartID,
		OwnerID: uuid.New(),
		Items: []CartItem{
			{ItemID: itemID, Quantity: 1}, // Item doesn't exist in store
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cartStore.carts[cartID] = cart

	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err == nil {
		t.Error("Expected error for item not found")
	}

	if presentation != nil {
		t.Error("Expected nil presentation for item not found")
	}
}

func TestCartPresentationRouter_getCartPresentation_NilCartStore(t *testing.T) {
	itemStore := NewMockCartPresentationItemStore()
	router := NewCartPresentationRouter(itemStore, nil)

	cartID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err == nil {
		t.Error("Expected error for nil cart store")
	}

	if presentation != nil {
		t.Error("Expected nil presentation for nil cart store")
	}
}

func TestCartPresentationRouter_getCartPresentation_NilItemStore(t *testing.T) {
	cartStore := NewMockCartPresentationCartStore()
	router := NewCartPresentationRouter(nil, cartStore)

	cartID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/presentation/cart/"+cartID.String(), nil)
	req.SetPathValue("id", cartID.String())

	presentation, err := router.getCartPresentation(context.Background(), req)

	if err == nil {
		t.Error("Expected error for nil item store")
	}

	if presentation != nil {
		t.Error("Expected nil presentation for nil item store")
	}
}
