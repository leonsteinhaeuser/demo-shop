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

type Cart struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OwnerID uuid.UUID  `json:"owner_id"`
	Items   []CartItem `json:"items"`
}

type CartItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

type CartStore interface {
	Create(ctx context.Context, cart *Cart) error
	Get(ctx context.Context, id uuid.UUID) (*Cart, error)
	Update(ctx context.Context, cart *Cart) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CartRouter struct {
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

	Store CartStore
}

func NewCartRouter(store CartStore) *CartRouter {
	return &CartRouter{
		processedCreateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_create_requests_total",
			Help: "Total number of cart create requests",
		}),
		processedCreateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_create_failures_total",
			Help: "Total number of cart create failures",
		}),
		processedUpdateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_update_requests_total",
			Help: "Total number of cart update requests",
		}),
		processedUpdateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_update_failures_total",
			Help: "Total number of cart update failures",
		}),
		processedDeleteRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_delete_requests_total",
			Help: "Total number of cart delete requests",
		}),
		processedDeleteFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_delete_failures_total",
			Help: "Total number of cart delete failures",
		}),
		processedGetRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_get_requests_total",
			Help: "Total number of cart get requests",
		}),
		processedGetFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_get_failures_total",
			Help: "Total number of cart get failures",
		}),
		processedListRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_list_requests_total",
			Help: "Total number of cart list requests",
		}),
		processedListFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cart_list_failures_total",
			Help: "Total number of cart list failures",
		}),
		Store: store,
	}
}

func (c *CartRouter) GetApiVersion() string {
	return version
}

func (c *CartRouter) GetGroup() string {
	return group
}

func (c *CartRouter) GetKind() string {
	return "carts"
}

func (c *CartRouter) Routes() []router.PathObject {
	return []router.PathObject{
		{
			Method: "POST",
			Func:   handlers.HttpPost(c.createCart),
		},
		{
			Path:   "/{id}",
			Method: "GET",
			Func:   handlers.HttpGet(c.getCart),
		},
		{
			Path:   "/{id}",
			Method: "PUT",
			Func:   handlers.HttpUpdate(c.updateCart),
		},
		{
			Path:   "/{id}",
			Method: "DELETE",
			Func:   handlers.HttpDelete(c.deleteCart),
		},
	}
}

func (c *CartRouter) createCart(ctx context.Context, r *http.Request, cart *Cart) error {
	c.processedCreateFailures.Inc()

	if c.Store == nil {
		c.processedCreateFailures.Inc()
		return errors.New("cart store is not initialized")
	}
	if cart.ID == uuid.Nil {
		cart.ID = uuid.New()
	}
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = cart.CreatedAt

	err := c.Store.Create(ctx, cart)
	if err != nil {
		c.processedCreateFailures.Inc()
		return err
	}
	return nil
}

func (c *CartRouter) getCart(ctx context.Context, r *http.Request) (*Cart, error) {
	c.processedGetRequests.Inc()

	if c.Store == nil {
		c.processedGetFailures.Inc()
		return nil, errors.New("cart store is not initialized")
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	cart, err := c.Store.Get(ctx, id)
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	return cart, nil
}

func (c *CartRouter) updateCart(ctx context.Context, r *http.Request, cart *Cart) error {
	c.processedUpdateRequests.Inc()

	if c.Store == nil {
		c.processedUpdateFailures.Inc()
		return errors.New("cart store is not initialized")
	}

	if cart.ID == uuid.Nil {
		c.processedUpdateFailures.Inc()
		return errors.New("cart ID cannot be empty")
	}

	cart.UpdatedAt = time.Now()

	err := c.Store.Update(ctx, cart)
	if err != nil {
		c.processedUpdateFailures.Inc()
		return err
	}
	return nil
}

func (c *CartRouter) deleteCart(ctx context.Context, r *http.Request, cart *Cart) error {
	c.processedDeleteRequests.Inc()

	if c.Store == nil {
		c.processedDeleteFailures.Inc()
		return errors.New("cart store is not initialized")
	}
	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		c.processedDeleteFailures.Inc()
		return err
	}
	if id != cart.ID {
		c.processedDeleteFailures.Inc()
		return errors.New("cart ID from path does not match cart ID in body")
	}

	err = c.Store.Delete(ctx, id)
	if err != nil {
		c.processedDeleteFailures.Inc()
		return err
	}
	return nil
}
