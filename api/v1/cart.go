package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
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
	Store CartStore
}

func (c *CartRouter) GetApiVersion() string {
	return "v1"
}

func (c *CartRouter) GetGroup() string {
	return "core"
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
			Func:   c.getCart,
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
	if c.Store == nil {
		return errors.New("cart store is not initialized")
	}

	if cart.ID == uuid.Nil {
		cart.ID = uuid.New()
	}
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = cart.CreatedAt

	return c.Store.Create(ctx, cart)
}

func (c *CartRouter) getCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusBadRequest,
			Path:    r.URL.Path,
			Message: "Invalid cart ID in path",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	cart, err := c.Store.Get(ctx, id)
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Path:    r.URL.Path,
			Message: "Failed to retrieve cart",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	// TODO: when url param format is "full", include full item details

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cart)
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Path:    r.URL.Path,
			Message: "Failed to encode cart",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}
}

func (c *CartRouter) updateCart(ctx context.Context, r *http.Request, cart *Cart) error {
	if c.Store == nil {
		return errors.New("cart store is not initialized")
	}

	if cart.ID == uuid.Nil {
		return errors.New("cart ID cannot be empty")
	}

	cart.UpdatedAt = time.Now()
	return c.Store.Update(ctx, cart)
}

func (c *CartRouter) deleteCart(ctx context.Context, r *http.Request, cart *Cart) error {
	if c.Store == nil {
		return errors.New("cart store is not initialized")
	}
	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return err
	}
	if id != cart.ID {
		return errors.New("cart ID from path does not match cart ID in body")
	}
	return c.Store.Delete(ctx, id)
}
