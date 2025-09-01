package utils

import (
	"fmt"
	"net/http"
)

func CheckHealth(address string) error {
	resp, err := http.Get(address)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %s", resp.Status)
	}
	return nil
}
