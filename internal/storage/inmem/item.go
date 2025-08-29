package inmem

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
)

var (
	_ apiv1.ItemStore = (*ItemInMemStorage)(nil)
)

type ItemInMemStorage struct {
	items map[string]*apiv1.Item
}

func NewItemInMemStorage() *ItemInMemStorage {
	itemApple := uuid.New()
	itemBanana := uuid.New()
	itemOrange := uuid.New()
	itemMango := uuid.New()

	return &ItemInMemStorage{
		items: map[string]*apiv1.Item{
			itemApple.String(): {
				ID:        itemApple,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),

				Name:        "Apple",
				Description: "A juicy red apple",
				Price:       0.75,
				Quantity:    200,
				Location:    "Aisle 1",
			},
			itemBanana.String(): {
				ID:        itemBanana,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),

				Name:        "Banana",
				Description: "A ripe yellow banana",
				Price:       1.99,
				Quantity:    150,
				Location:    "Aisle 1",
			},
			itemOrange.String(): {
				ID:        itemOrange,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),

				Name:        "Orange",
				Description: "A sweet orange",
				Price:       3.00,
				Quantity:    100,
				Location:    "Aisle 1",
			},
			itemMango.String(): {
				ID:        itemMango,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),

				Name:        "Mango",
				Description: "A ripe mango",
				Price:       4.00,
				Quantity:    100,
				Location:    "Aisle 1",
			},
		},
	}
}

func (i *ItemInMemStorage) Create(ctx context.Context, item *apiv1.Item) error {
	// create unique item id
	for {
		id := uuid.New()
		if _, exists := i.items[id.String()]; exists {
			continue
		}
		item.ID = id
		break
	}
	i.items[item.ID.String()] = item
	return nil
}

func (i *ItemInMemStorage) List(ctx context.Context, page, limit int) ([]apiv1.Item, error) {
	var items []apiv1.Item
	for _, item := range i.items {
		items = append(items, *item)
	}
	return items, nil
}

func (i *ItemInMemStorage) Get(ctx context.Context, id uuid.UUID) (*apiv1.Item, error) {
	item, exists := i.items[id.String()]
	if !exists {
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (i *ItemInMemStorage) Update(ctx context.Context, item *apiv1.Item) error {
	i.items[item.ID.String()] = item
	return nil
}

func (i *ItemInMemStorage) Delete(ctx context.Context, id uuid.UUID) error {
	delete(i.items, id.String())
	return nil
}
