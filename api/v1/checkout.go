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
	Store CheckoutStore
}

func (c *CheckoutRouter) GetApiVersion() string {
	return "v1"
}

func (c *CheckoutRouter) GetGroup() string {
	return "core"
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
	if checkout.UserID == uuid.Nil {
		return errors.New("UserID cannot be nil")
	}
	if checkout.CartID == uuid.Nil {
		return errors.New("CartID cannot be nil")
	}
	return c.Store.Create(ctx, checkout)
}

func (c *CheckoutRouter) getCheckout(ctx context.Context, r *http.Request) (*Checkout, error) {
	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return nil, err
	}
	return c.Store.Get(ctx, id)
}

func (c *CheckoutRouter) updateCheckout(ctx context.Context, r *http.Request, checkout *Checkout) error {
	return c.Store.Update(ctx, checkout)
}

func (c *CheckoutRouter) deleteCheckout(ctx context.Context, r *http.Request, checkout *Checkout) error {
	return c.Store.Delete(ctx, checkout.ID)
}
