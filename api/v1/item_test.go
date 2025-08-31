package v1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
)

// MockItemStore implements ItemStore interface for testing
type MockItemStore struct {
	items       map[uuid.UUID]*Item
	shouldError bool
}

func NewMockItemStore() *MockItemStore {
	return &MockItemStore{
		items: make(map[uuid.UUID]*Item),
	}
}

func (m *MockItemStore) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockItemStore) Create(ctx context.Context, item *Item) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockItemStore) Get(ctx context.Context, id uuid.UUID) (*Item, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	item, exists := m.items[id]
	if !exists {
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (m *MockItemStore) Update(ctx context.Context, item *Item) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockItemStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	delete(m.items, id)
	return nil
}

func (m *MockItemStore) List(ctx context.Context, page, limit int) ([]Item, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	items := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		items = append(items, *item)
	}
	return items, nil
}

func TestNewItemRouter(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	if router == nil {
		t.Error("Expected router to be created")
	}
}

func TestItemRouter_GetApiVersion(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	version := router.GetApiVersion()
	expected := "v1"

	if version != expected {
		t.Errorf("Expected version %s, got %s", expected, version)
	}
}

func TestItemRouter_GetGroup(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	group := router.GetGroup()
	expected := "core"

	if group != expected {
		t.Errorf("Expected group %s, got %s", expected, group)
	}
}

func TestItemRouter_GetKind(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	kind := router.GetKind()
	expected := "items"

	if kind != expected {
		t.Errorf("Expected kind %s, got %s", expected, kind)
	}
}

func TestItemRouter_createItem_Success(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.Nil, // ID should be empty for creation
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify item was stored (ID should now be set)
	if item.ID == uuid.Nil {
		t.Error("Expected ID to be set after creation")
	}

	storedItem, exists := store.items[item.ID]
	if !exists {
		t.Error("Expected item to be stored")
	}

	if storedItem.Name != item.Name {
		t.Errorf("Expected name %s, got %s", item.Name, storedItem.Name)
	}

	if storedItem.Price != item.Price {
		t.Errorf("Expected price %f, got %f", item.Price, storedItem.Price)
	}
}

func TestItemRouter_createItem_NilStore(t *testing.T) {
	router := NewItemRouter(nil)

	item := &Item{
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error for nil store")
	}
}

func TestItemRouter_createItem_NilItem(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error for nil item")
	}
}

func TestItemRouter_createItem_NonEmptyID(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.New(), // Non-empty ID should cause error
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error for non-empty ID")
	}
}

func TestItemRouter_createItem_EmptyName(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.Nil,
		Name:        "", // Empty name should cause error
		Description: "A test item",
		Price:       19.99,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestItemRouter_createItem_ZeroPrice(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.Nil,
		Name:        "Test Item",
		Description: "A test item",
		Price:       0, // Zero price should cause error
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error for zero price")
	}
}

func TestItemRouter_createItem_NegativePrice(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.Nil,
		Name:        "Test Item",
		Description: "A test item",
		Price:       -10.0, // Negative price should cause error
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error for negative price")
	}
}

func TestItemRouter_createItem_StoreError(t *testing.T) {
	store := NewMockItemStore()
	store.SetError(true)
	router := NewItemRouter(store)

	item := &Item{
		ID:          uuid.Nil,
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/items", nil)
	err := router.createItem(context.Background(), req, item)

	if err == nil {
		t.Error("Expected error from store")
	}
}

func TestItemRouter_getItem_Success(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	itemID := uuid.New()
	item := &Item{
		ID:          itemID,
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	store.items[itemID] = item

	req := httptest.NewRequest("GET", "/api/v1/core/items/"+itemID.String(), nil)
	req.SetPathValue("id", itemID.String())

	retrievedItem, err := router.getItem(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedItem == nil {
		t.Fatal("Expected item to be returned")
	}

	if retrievedItem.ID != itemID {
		t.Errorf("Expected item ID %s, got %s", itemID, retrievedItem.ID)
	}

	if retrievedItem.Name != "Test Item" {
		t.Errorf("Expected name Test Item, got %s", retrievedItem.Name)
	}
}

func TestItemRouter_getItem_NotFound(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	itemID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/core/items/"+itemID.String(), nil)
	req.SetPathValue("id", itemID.String())

	item, err := router.getItem(context.Background(), req)

	if err == nil {
		t.Error("Expected error for non-existent item")
	}

	if item != nil {
		t.Error("Expected nil item for non-existent ID")
	}
}

func TestItemRouter_updateItem_Success(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	itemID := uuid.New()
	originalItem := &Item{
		ID:          itemID,
		Name:        "Original Item",
		Description: "Original description",
		Price:       19.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	store.items[itemID] = originalItem

	updatedItem := &Item{
		ID:          itemID,
		Name:        "Updated Item",
		Description: "Updated description",
		Price:       29.99,
	}

	req := httptest.NewRequest("PUT", "/api/v1/core/items/"+itemID.String(), nil)
	req.SetPathValue("id", itemID.String())

	err := router.updateItem(context.Background(), req, updatedItem)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify item was updated
	storedItem := store.items[itemID]
	if storedItem.Name != "Updated Item" {
		t.Errorf("Expected updated name Updated Item, got %s", storedItem.Name)
	}

	if storedItem.Price != 29.99 {
		t.Errorf("Expected updated price 29.99, got %f", storedItem.Price)
	}
}

func TestItemRouter_deleteItem_Success(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	itemID := uuid.New()
	item := &Item{
		ID:          itemID,
		Name:        "Test Item",
		Description: "A test item",
		Price:       19.99,
	}
	store.items[itemID] = item

	req := httptest.NewRequest("DELETE", "/api/v1/core/items/"+itemID.String(), nil)
	req.SetPathValue("id", itemID.String())

	err := router.deleteItem(context.Background(), req, item)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify item was deleted
	_, exists := store.items[itemID]
	if exists {
		t.Error("Expected item to be deleted")
	}
}

func TestItemRouter_listItems_Success(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	// Add test items
	item1 := &Item{
		ID:          uuid.New(),
		Name:        "Item 1",
		Description: "First item",
		Price:       19.99,
	}
	item2 := &Item{
		ID:          uuid.New(),
		Name:        "Item 2",
		Description: "Second item",
		Price:       29.99,
	}

	store.items[item1.ID] = item1
	store.items[item2.ID] = item2

	req := httptest.NewRequest("GET", "/api/v1/core/items", nil)
	filters := handlers.FilterObjectList{Page: 0, Limit: 10}

	items, err := router.listItems(context.Background(), req, filters)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestItemRouter_listItems_Empty(t *testing.T) {
	store := NewMockItemStore()
	router := NewItemRouter(store)

	req := httptest.NewRequest("GET", "/api/v1/core/items", nil)
	filters := handlers.FilterObjectList{Page: 0, Limit: 10}

	items, err := router.listItems(context.Background(), req, filters)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}
