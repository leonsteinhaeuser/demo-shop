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

type Checkout struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	CartID    uuid.UUID `json:"cart_id"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"` // e.g., "pending", "completed", "failed"
}

type CheckoutStore interface {
	Create(ctx context.Context, checkout *Checkout) error
	Get(ctx context.Context, id uuid.UUID) (*Checkout, error)
	Update(ctx context.Context, checkout *Checkout) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CheckoutRouter struct {
	processedCreateRequests prometheus.Counter
	processedCreateFailures prometheus.Counter
	processedUpdateRequests prometheus.Counter
	processedUpdateFailures prometheus.Counter
	processedDeleteRequests prometheus.Counter
	processedDeleteFailures prometheus.Counter
	processedGetRequests    prometheus.Counter
	processedGetFailures    prometheus.Counter

	Store CheckoutStore
}

func NewCheckoutRouter(store CheckoutStore) *CheckoutRouter {
	return &CheckoutRouter{
		processedCreateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_create_requests_total",
			Help: "Total number of checkout create requests",
		}),
		processedCreateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_create_failures_total",
			Help: "Total number of checkout create failures",
		}),
		processedUpdateRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_update_requests_total",
			Help: "Total number of checkout update requests",
		}),
		processedUpdateFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_update_failures_total",
			Help: "Total number of checkout update failures",
		}),
		processedDeleteRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_delete_requests_total",
			Help: "Total number of checkout delete requests",
		}),
		processedDeleteFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_delete_failures_total",
			Help: "Total number of checkout delete failures",
		}),
		processedGetRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_get_requests_total",
			Help: "Total number of checkout get requests",
		}),
		processedGetFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "checkout_get_failures_total",
			Help: "Total number of checkout get failures",
		}),
		Store: store,
	}
}

func (c *CheckoutRouter) GetApiVersion() string {
	return version
}

func (c *CheckoutRouter) GetGroup() string {
	return group
}

func (c *CheckoutRouter) GetKind() string {
	return "checkouts"
}

func (c *CheckoutRouter) Routes() []router.PathObject {
	return []router.PathObject{
		{
			Method: "POST",
			Func:   handlers.HttpPost(c.createCheckout),
		},
		{
			Method: "GET",
			Func:   handlers.HttpGet(c.getCheckout),
		},
		{
			Method: "PUT",
			Func:   handlers.HttpUpdate(c.updateCheckout),
		},
		{
			Method: "DELETE",
			Func:   handlers.HttpDelete(c.deleteCheckout),
		},
	}
}

func (c *CheckoutRouter) createCheckout(ctx context.Context, r *http.Request, checkout *Checkout) error {
	c.processedCreateRequests.Inc()

	if checkout.UserID == uuid.Nil {
		c.processedCreateFailures.Inc()
		return errors.New("UserID cannot be nil")
	}
	if checkout.CartID == uuid.Nil {
		c.processedCreateFailures.Inc()
		return errors.New("CartID cannot be nil")
	}

	err := c.Store.Create(ctx, checkout)
	if err != nil {
		c.processedCreateFailures.Inc()
		return err
	}
	return nil
}

func (c *CheckoutRouter) getCheckout(ctx context.Context, r *http.Request) (*Checkout, error) {
	c.processedGetRequests.Inc()

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	checkout, err := c.Store.Get(ctx, id)
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	return checkout, nil
}

func (c *CheckoutRouter) updateCheckout(ctx context.Context, r *http.Request, checkout *Checkout) error {
	c.processedUpdateRequests.Inc()

	err := c.Store.Update(ctx, checkout)
	if err != nil {
		c.processedUpdateFailures.Inc()
		return err
	}
	return nil
}

func (c *CheckoutRouter) deleteCheckout(ctx context.Context, r *http.Request, checkout *Checkout) error {
	c.processedDeleteRequests.Inc()

	if checkout == nil {
		c.processedDeleteFailures.Inc()
		return errors.New("checkout cannot be nil")
	}

	err := c.Store.Delete(ctx, checkout.ID)
	if err != nil {
		c.processedDeleteFailures.Inc()
		return err
	}
	return nil
}
