package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/prometheus/client_golang/prometheus"
)

type CartPresentation struct {
	Items      []CartItemPresentation `json:"items"`
	TotalPrice float64                `json:"total_price"`
}

type CartItemPresentation struct {
	Item       Item    `json:"item"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
}

type CartPresentationRouter struct {
	ItemStore            ItemStore
	CartStore            CartStore
	processedGetRequests prometheus.Counter
	processedGetFailures prometheus.Counter
}

func NewCartPresentationRouter(itemStore ItemStore, cartStore CartStore) *CartPresentationRouter {
	return &CartPresentationRouter{
		ItemStore: itemStore,
		CartStore: cartStore,
		processedGetRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cartpresentation_get_processed_requests_total",
				Help: "Total number of cart presentation get requests",
			},
		),
		processedGetFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cartpresentation_get_processed_failures_total",
				Help: "Total number of cart presentation get request failures",
			},
		),
	}
}

func (c *CartPresentationRouter) GetApiVersion() string {
	return version
}

func (c *CartPresentationRouter) GetGroup() string {
	return "presentation"
}

func (c *CartPresentationRouter) GetKind() string {
	return "cart"
}

func (c *CartPresentationRouter) Routes() []router.PathObject {
	return []router.PathObject{
		{
			Path:   "/{id}",
			Method: "GET",
			Func:   handlers.HttpGet(c.getCartPresentation),
		},
	}
}

func (c *CartPresentationRouter) getCartPresentation(ctx context.Context, r *http.Request) (*CartPresentation, error) {
	c.processedGetRequests.Inc()

	if c.CartStore == nil {
		c.processedGetFailures.Inc()
		return nil, errors.New("cart store is not initialized")
	}
	if c.ItemStore == nil {
		c.processedGetFailures.Inc()
		return nil, errors.New("item store is not initialized")
	}

	cartID, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	cart, err := c.CartStore.Get(ctx, cartID)
	if err != nil {
		c.processedGetFailures.Inc()
		return nil, err
	}

	if cart == nil {
		c.processedGetFailures.Inc()
		return nil, errors.New("cart not found")
	}

	if len(cart.Items) == 0 {
		return &CartPresentation{Items: []CartItemPresentation{}, TotalPrice: 0.0}, nil
	}

	cp := &CartPresentation{Items: []CartItemPresentation{}, TotalPrice: 0.0}

	// TODO: this can be optimized to fetch all items in multiple goroutines
	// retrieve item details for each cart item
	for _, cartItem := range cart.Items {
		item, err := c.ItemStore.Get(ctx, cartItem.ItemID)
		if err != nil {
			c.processedGetFailures.Inc()
			return nil, err
		}
		if item == nil {
			c.processedGetFailures.Inc()
			return nil, errors.New("item not found for cart item")
		}
		// create CartItemPresentation
		cartItemPresentation := CartItemPresentation{
			Item:       *item,
			Quantity:   cartItem.Quantity,
			TotalPrice: item.Price * float64(cartItem.Quantity),
		}
		cp.Items = append(cp.Items, cartItemPresentation)
		cp.TotalPrice += cartItemPresentation.TotalPrice
	}
	return cp, nil
}
