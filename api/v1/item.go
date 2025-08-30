package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/prometheus/client_golang/prometheus"
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
	processedCreateRequests prometheus.Counter
	processedCreateFailures prometheus.Counter
	processedUpdateRequests prometheus.Counter
	processedUpdateFailures prometheus.Counter
	processedDeleteRequests prometheus.Counter
	processedDeleteFailures prometheus.Counter
	processedGetRequests    prometheus.Counter
	processedGetFailures    prometheus.Counter
	processedListRequests   prometheus.Counter
	processedListFailures   prometheus.Counter

	Store ItemStore
}

func NewItemRouter(store ItemStore) *ItemRouter {
	return &ItemRouter{
		processedCreateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_create_requests_total",
			Help: "Total number of item create requests",
		}),
		processedCreateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_create_failures_total",
			Help: "Total number of item create failures",
		}),
		processedUpdateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_update_requests_total",
			Help: "Total number of item update requests",
		}),
		processedUpdateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_update_failures_total",
			Help: "Total number of item update failures",
		}),
		processedDeleteRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_delete_requests_total",
			Help: "Total number of item delete requests",
		}),
		processedDeleteFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_delete_failures_total",
			Help: "Total number of item delete failures",
		}),
		processedGetRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_get_requests_total",
			Help: "Total number of item get requests",
		}),
		processedGetFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_get_failures_total",
			Help: "Total number of item get failures",
		}),
		processedListRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_list_requests_total",
			Help: "Total number of item list requests",
		}),
		processedListFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "item_list_failures_total",
			Help: "Total number of item list failures",
		}),
		Store: store,
	}
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
	i.processedCreateRequests.Inc()

	if i.Store == nil {
		i.processedCreateFailures.Inc()
		return errors.New("item store is not initialized")
	}
	if item == nil {
		i.processedCreateFailures.Inc()
		return errors.New("item cannot be nil")
	}
	if item.ID != uuid.Nil {
		i.processedCreateFailures.Inc()
		return errors.New("item ID must be empty for creation")
	}
	if item.Name == "" {
		i.processedCreateFailures.Inc()
		return errors.New("item name cannot be empty")
	}
	if item.Price <= 0 {
		i.processedCreateFailures.Inc()
		return errors.New("item price must be greater than zero")
	}
	item.ID = uuid.New()
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt

	err := i.Store.Create(ctx, item)
	if err != nil {
		i.processedCreateFailures.Inc()
		return err
	}
	return nil
}

func (i *ItemRouter) listItems(ctx context.Context, r *http.Request, filters handlers.FilterObjectList) ([]Item, error) {
	i.processedListRequests.Inc()

	if i.Store == nil {
		i.processedListFailures.Inc()
		return nil, errors.New("item store is not initialized")
	}

	items, err := i.Store.List(ctx, filters.Page, filters.Limit)
	if err != nil {
		i.processedListFailures.Inc()
		return nil, err
	}
	return items, nil
}

func (i *ItemRouter) getItem(ctx context.Context, r *http.Request) (*Item, error) {
	i.processedGetRequests.Inc()

	if i.Store == nil {
		i.processedGetFailures.Inc()
		return nil, errors.New("item store is not initialized")
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		i.processedGetFailures.Inc()
		return nil, err
	}

	item, err := i.Store.Get(ctx, id)
	if err != nil {
		i.processedGetFailures.Inc()
		return nil, err
	}
	if item == nil {
		i.processedGetFailures.Inc()
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (i *ItemRouter) updateItem(ctx context.Context, r *http.Request, item *Item) error {
	i.processedUpdateRequests.Inc()

	if i.Store == nil {
		i.processedUpdateFailures.Inc()
		return errors.New("item store is not initialized")
	}

	if item.ID == uuid.Nil {
		i.processedUpdateFailures.Inc()
		return errors.New("item ID cannot be empty")
	}
	if item.Name == "" {
		i.processedUpdateFailures.Inc()
		return errors.New("item name cannot be empty")
	}
	if item.Price <= 0 {
		i.processedUpdateFailures.Inc()
		return errors.New("item price must be greater than zero")
	}

	err := i.Store.Update(ctx, item)
	if err != nil {
		i.processedUpdateFailures.Inc()
		return err
	}
	return nil
}

func (i *ItemRouter) deleteItem(ctx context.Context, r *http.Request, item *Item) error {
	i.processedDeleteRequests.Inc()

	if i.Store == nil {
		i.processedDeleteFailures.Inc()
		return errors.New("item store is not initialized")
	}

	if item == nil {
		i.processedDeleteFailures.Inc()
		return errors.New("item is nil")
	}

	err := i.Store.Delete(ctx, item.ID)
	if err != nil {
		i.processedDeleteFailures.Inc()
		return err
	}
	return nil
}
