package check

import (
	"context"
	"log/slog"
	"time"

	"github.com/leonsteinhaeuser/demo-shop/internal/utils"
)

const (
	livenessPath = "/health/liveness"
)

func Check(ctx context.Context, userServiceURL, cartServiceURL, itemServiceURL, checkoutServiceURL, cartPresentationServiceURL string) {
	processChecks(ctx, userServiceURL+livenessPath)
	processChecks(ctx, cartServiceURL+livenessPath)
	processChecks(ctx, itemServiceURL+livenessPath)
	processChecks(ctx, checkoutServiceURL+livenessPath)
	processChecks(ctx, cartPresentationServiceURL+livenessPath)
}

func processChecks(ctx context.Context, url string) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := utils.CheckHealth(url + livenessPath); err != nil {
					slog.Error("Ping check failed", "url", url, "error", err)
				}
			}
		}
	}()
}
