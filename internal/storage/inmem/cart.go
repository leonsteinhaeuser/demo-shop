package inmem

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

var (
	_ apiv1.CartStore = (*CartInMemStorage)(nil)
)

type CartInMemStorage struct {
	carts map[string]*apiv1.Cart
}

func NewCartInMemStorage() *CartInMemStorage {
	defaultCart := uuid.New()

	return &CartInMemStorage{
		carts: map[string]*apiv1.Cart{
			defaultCart.String(): {
				ID:        defaultCart,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				OwnerID:   defaultUser,
				Items:     []apiv1.CartItem{},
			},
		},
	}
}

func (c *CartInMemStorage) Create(ctx context.Context, cart *apiv1.Cart) error {
	for {
		id := uuid.New()
		if _, exists := c.carts[id.String()]; exists {
			continue
		}
		cart.ID = id
		break
	}
	c.carts[cart.ID.String()] = cart
	return nil
}

func (c *CartInMemStorage) Get(ctx context.Context, id uuid.UUID) (*apiv1.Cart, error) {
	cart, exists := c.carts[id.String()]
	if !exists {
		return nil, errors.New("cart not found")
	}
	return cart, nil
}

func (c *CartInMemStorage) Update(ctx context.Context, cart *apiv1.Cart) error {
	c.carts[cart.ID.String()] = cart
	return nil
}

func (c *CartInMemStorage) Delete(ctx context.Context, id uuid.UUID) error {
	delete(c.carts, id.String())
	return nil
}
