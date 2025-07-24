package v1

import (
	"testing"
)

func TestClientURLGeneration(t *testing.T) {
	baseURL := "http://localhost:8080"

	// Test cart client URL generation
	cartClient := NewCartClient(baseURL)
	if cartClient.baseURL != baseURL {
		t.Errorf("Expected cart client baseURL %s, got %s", baseURL, cartClient.baseURL)
	}

	// Test item client URL generation
	itemClient := NewItemClient(baseURL)
	if itemClient.baseURL != baseURL {
		t.Errorf("Expected item client baseURL %s, got %s", baseURL, itemClient.baseURL)
	}

	// Test user client URL generation
	userClient := NewUserClient(baseURL)
	if userClient.baseURL != baseURL {
		t.Errorf("Expected user client baseURL %s, got %s", baseURL, userClient.baseURL)
	}

	// Test cart presentation client URL generation
	cartPresentationClient := NewCartPresentationClient(baseURL)
	if cartPresentationClient.baseURL != baseURL {
		t.Errorf("Expected cart presentation client baseURL %s, got %s", baseURL, cartPresentationClient.baseURL)
	}
}

func TestClientsFactory(t *testing.T) {
	baseURL := "https://api.example.com"

	clients := NewDefaultClients(baseURL)

	if clients.Cart == nil {
		t.Error("Expected Cart client to be initialized")
	}

	if clients.Item == nil {
		t.Error("Expected Item client to be initialized")
	}

	if clients.User == nil {
		t.Error("Expected User client to be initialized")
	}

	if clients.CartPresentation == nil {
		t.Error("Expected CartPresentation client to be initialized")
	}
}
