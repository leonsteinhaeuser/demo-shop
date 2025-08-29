package inmem

import (
	"context"
	"errors"

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
				ID:    itemApple,
				Name:  "Apple",
				Price: 0.75,
			},
			itemBanana.String(): {
				ID:    itemBanana,
				Name:  "Banana",
				Price: 1.99,
			},
			itemOrange.String(): {
				ID:    itemOrange,
				Name:  "Orange",
				Price: 3.00,
			},
			itemMango.String(): {
				ID:    itemMango,
				Name:  "Mango",
				Price: 4.00,
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
