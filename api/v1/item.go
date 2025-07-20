package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

var (
	_ router.ApiObject = &ItemRouter{}
)

type Item struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Location    string  `json:"location"`
}

type ItemStore interface {
	Create(ctx context.Context, item *Item) error
	List(ctx context.Context, page, limit int) ([]Item, error)
	Get(ctx context.Context, id uuid.UUID) (*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ItemRouter struct {
	Store ItemStore
}

func (i *ItemRouter) GetApiVersion() string {
	return "v1"
}

func (i *ItemRouter) GetGroup() string {
	return "core"
}

func (i *ItemRouter) GetKind() string {
	return "items"
}

func (i *ItemRouter) Routes() []router.PathObject {
	return []router.PathObject{
		{
			Method: "POST",
			Func:   handlers.HttpPost(i.createItem),
		},
		{
			Method: "GET",
			Func:   handlers.HttpList(i.listItems),
		},
		{
			Path:   "/{id}",
			Method: "GET",
			Func:   handlers.HttpGet(i.getItem),
		},
		{
			Path:   "/{id}",
			Method: "PUT",
			Func:   handlers.HttpUpdate(i.updateItem),
		},
		{
			Path:   "/{id}",
			Method: "DELETE",
			Func:   handlers.HttpDelete(i.deleteItem),
		},
	}
}

func (i *ItemRouter) createItem(ctx context.Context, r *http.Request, item *Item) error {
	if i.Store == nil {
		return errors.New("item store is not initialized")
	}
	if item == nil {
		return errors.New("item cannot be nil")
	}
	if item.ID != uuid.Nil {
		return errors.New("item ID must be empty for creation")
	}
	if item.Name == "" {
		return errors.New("item name cannot be empty")
	}
	if item.Price <= 0 {
		return errors.New("item price must be greater than zero")
	}
	item.ID = uuid.New()
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt
	if err := i.Store.Create(ctx, item); err != nil {
		return err
	}
	return nil
}

func (i *ItemRouter) listItems(ctx context.Context, r *http.Request, filters handlers.FilterObjectList) ([]Item, error) {
	if i.Store == nil {
		return nil, errors.New("item store is not initialized")
	}
	return i.Store.List(ctx, filters.Page, filters.Limit)
}

func (i *ItemRouter) getItem(ctx context.Context, r *http.Request) (*Item, error) {
	if i.Store == nil {
		return nil, errors.New("item store is not initialized")
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return nil, err
	}

	item, err := i.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (i *ItemRouter) updateItem(ctx context.Context, r *http.Request, item *Item) error {
	if i.Store == nil {
		return errors.New("item store is not initialized")
	}

	if item.ID == uuid.Nil {
		return errors.New("item ID cannot be empty")
	}
	if item.Name == "" {
		return errors.New("item name cannot be empty")
	}
	if item.Price <= 0 {
		return errors.New("item price must be greater than zero")
	}
	if err := i.Store.Update(ctx, item); err != nil {
		return err
	}
	return nil
}

func (i *ItemRouter) deleteItem(ctx context.Context, r *http.Request, item *Item) error {
	if i.Store == nil {
		return errors.New("item store is not initialized")
	}

	if item == nil {
		return errors.New("item is nil")
	}
	if err := i.Store.Delete(ctx, item.ID); err != nil {
		return err
	}
	return nil
}
