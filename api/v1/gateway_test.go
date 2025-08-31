package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	cookieEncryptionKey []byte = []byte("a_random_secret_key")
)

func TestGateway_HandleLogin(t *testing.T) {
	gateway := NewGateway(
		"http://localhost:8084", // userServiceURL
		"http://localhost:8082", // cartServiceURL
		"http://localhost:8081", // itemServiceURL
		"http://localhost:8085", // checkoutServiceURL
		"http://localhost:8083", // cartPresentationServiceURL
		cookieEncryptionKey,
	)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectCookie   bool
	}{
		{
			name: "valid login request",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized, // Will fail because user doesn't exist
			expectCookie:   false,
		},
		{
			name: "missing email",
			requestBody: LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
		},
		{
			name: "missing password",
			requestBody: LoginRequest{
				Username: "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			gateway.handleLogin(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			cookies := rr.Result().Cookies()
			hasCookie := len(cookies) > 0 && cookies[0].Name == "session"

			if tt.expectCookie && !hasCookie {
				t.Error("Expected session cookie to be set")
			}

			if !tt.expectCookie && hasCookie {
				t.Error("Did not expect session cookie to be set")
			}
		})
	}
}

func TestGateway_HandleLogout(t *testing.T) {
	gateway := NewGateway(
		"http://localhost:8084", // userServiceURL
		"http://localhost:8082", // cartServiceURL
		"http://localhost:8081", // itemServiceURL
		"http://localhost:8085", // checkoutServiceURL
		"http://localhost:8083", // cartPresentationServiceURL
		cookieEncryptionKey,
	)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()

	gateway.handleLogout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that logout clears the session cookie
	cookies := rr.Result().Cookies()
	if len(cookies) > 0 {
		cookie := cookies[0]
		if cookie.Name == "session" && cookie.MaxAge != -1 {
			t.Error("Expected session cookie to be cleared (MaxAge = -1)")
		}
	}
}

func TestGateway_ServeHTTP(t *testing.T) {
	gateway := NewGateway(
		"http://localhost:8084", // userServiceURL
		"http://localhost:8082", // cartServiceURL
		"http://localhost:8081", // itemServiceURL
		"http://localhost:8085", // checkoutServiceURL
		"http://localhost:8083", // cartPresentationServiceURL
		cookieEncryptionKey,
	)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "login endpoint POST",
			path:           "/login",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // No body provided
		},
		{
			name:           "logout endpoint POST",
			path:           "/logout",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "login endpoint GET not allowed",
			path:           "/login",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unknown endpoint",
			path:           "/unknown",
			method:         http.MethodPost,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			gateway.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGateway_ValidatePassword(t *testing.T) {
	gateway := NewGateway(
		"http://localhost:8084", // userServiceURL
		"http://localhost:8082", // cartServiceURL
		"http://localhost:8081", // itemServiceURL
		"http://localhost:8085", // checkoutServiceURL
		"http://localhost:8083", // cartPresentationServiceURL
		cookieEncryptionKey,
	)

	tests := []struct {
		name           string
		password       string
		hashedPassword *string
		expected       bool
	}{
		{
			name:           "valid plain text password",
			password:       "password123",
			hashedPassword: &[]string{"password123"}[0],
			expected:       true,
		},
		{
			name:           "invalid password",
			password:       "wrongpassword",
			hashedPassword: &[]string{"password123"}[0],
			expected:       false,
		},
		{
			name:           "nil hashed password",
			password:       "password123",
			hashedPassword: nil,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gateway.validatePassword(tt.password, tt.hashedPassword)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
