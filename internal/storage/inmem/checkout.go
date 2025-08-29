package inmem

import (
	"context"
	"errors"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	v1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

var (
	_ v1.CheckoutStore = (*CheckoutInMemStorage)(nil)
)

type CheckoutInMemStorage struct {
	checkouts map[string]*apiv1.Checkout
}

func NewCheckoutInMemStorage() *CheckoutInMemStorage {
	return &CheckoutInMemStorage{
		checkouts: map[string]*apiv1.Checkout{},
	}
}

func (c *CheckoutInMemStorage) Create(ctx context.Context, checkout *apiv1.Checkout) error {
	for {
		id := uuid.New()
		if _, exists := c.checkouts[id.String()]; exists {
			continue
		}
		checkout.ID = id
		break
	}
	c.checkouts[checkout.ID.String()] = checkout
	return nil
}

func (c *CheckoutInMemStorage) Get(ctx context.Context, id uuid.UUID) (*apiv1.Checkout, error) {
	checkout, exists := c.checkouts[id.String()]
	if !exists {
		return nil, errors.New("checkout not found")
	}
	return checkout, nil
}

func (c *CheckoutInMemStorage) Update(ctx context.Context, checkout *apiv1.Checkout) error {
	c.checkouts[checkout.ID.String()] = checkout
	return nil
}

func (c *CheckoutInMemStorage) Delete(ctx context.Context, id uuid.UUID) error {
	delete(c.checkouts, id.String())
	return nil
}
