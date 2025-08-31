package v1

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

const (
	gatewayVersion = "v1"
	gatewayGroup   = "gateway"
)

// LoginRequest represents the authentication request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the authentication response
type LoginResponse struct {
	User   User   `json:"user"`
	CartID string `json:"cart_id"`
}

// SessionData represents the data stored in the secure cookie
type SessionData struct {
	UserID   string `json:"user_id"`
	CartID   string `json:"cart_id"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

// Gateway handles authentication and request proxying
type Gateway struct {
	userServiceURL             string
	cartServiceURL             string
	itemServiceURL             string
	checkoutServiceURL         string
	cartPresentationServiceURL string
	cookieKey                  []byte
}

// NewGateway creates a new gateway instance
func NewGateway(userServiceURL, cartServiceURL, itemServiceURL, checkoutServiceURL, cartPresentationServiceURL string) *Gateway {
	// Generate a random cookie encryption key (in production, use a fixed key from config)
	cookieKey := make([]byte, 32)
	rand.Read(cookieKey)

	return &Gateway{
		userServiceURL:             userServiceURL,
		cartServiceURL:             cartServiceURL,
		itemServiceURL:             itemServiceURL,
		checkoutServiceURL:         checkoutServiceURL,
		cartPresentationServiceURL: cartPresentationServiceURL,
		cookieKey:                  cookieKey,
	}
}

func (g *Gateway) GetApiVersion() string {
	return gatewayVersion
}

func (g *Gateway) GetGroup() string {
	return gatewayGroup
}

func (g *Gateway) GetKind() string {
	return "auth"
}

func (g *Gateway) RegisterRoutes(mux *http.ServeMux) {
	// Authentication routes
	authPattern := fmt.Sprintf("/api/%s/auth/", g.GetApiVersion())
	mux.Handle(authPattern, http.StripPrefix(authPattern[:len(authPattern)-1], g))

	// Proxy routes for other services
	mux.HandleFunc("/api/v1/core/users", g.proxyToService)
	mux.HandleFunc("/api/v1/core/users/", g.proxyToService)
	mux.HandleFunc("/api/v1/core/carts", g.proxyToService)
	mux.HandleFunc("/api/v1/core/carts/", g.proxyToService)
	mux.HandleFunc("/api/v1/core/items", g.proxyToService)
	mux.HandleFunc("/api/v1/core/items/", g.proxyToService)
	mux.HandleFunc("/api/v1/core/checkouts", g.proxyToService)
	mux.HandleFunc("/api/v1/core/checkouts/", g.proxyToService)
	mux.HandleFunc("/api/v1/presentation/cart", g.proxyToService)
	mux.HandleFunc("/api/v1/presentation/cart/", g.proxyToService)
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/login" && r.Method == http.MethodPost:
		g.handleLogin(w, r)
	case r.URL.Path == "/logout" && r.Method == http.MethodPost:
		g.handleLogout(w, r)
	case r.URL.Path == "/api/v1/auth/login" && r.Method == http.MethodPost:
		g.handleLogin(w, r)
	case r.URL.Path == "/api/v1/auth/logout" && r.Method == http.MethodPost:
		g.handleLogout(w, r)
	default:
		(&router.ErrorResponse{
			Status:  http.StatusNotFound,
			Path:    r.URL.Path,
			Message: "endpoint not found",
		}).WriteTo(w)
	}
}

// handleLogin processes authentication requests
func (g *Gateway) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusBadRequest,
			Path:    r.URL.Path,
			Message: "invalid request body",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		(&router.ErrorResponse{
			Status:  http.StatusBadRequest,
			Path:    r.URL.Path,
			Message: "username and password are required",
		}).WriteTo(w)
		return
	}

	// Get user from user service
	userWithPassword, err := g.getUserByUsername(loginReq.Username)
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Path:    r.URL.Path,
			Message: "invalid credentials",
		}).WriteTo(w)
		return
	}

	// Validate password
	if !g.validatePassword(loginReq.Password, userWithPassword.Password) {
		(&router.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Path:    r.URL.Path,
			Message: "invalid credentials",
		}).WriteTo(w)
		return
	}

	// Create or get cart for user
	cartID, err := g.getOrCreateCartForUser(userWithPassword.User.ID.String())
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Path:    r.URL.Path,
			Message: "failed to create cart",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	// Create session data
	sessionData := SessionData{
		UserID:   userWithPassword.User.ID.String(),
		CartID:   cartID,
		Username: *userWithPassword.User.Username,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create secure cookie
	if err := g.setSessionCookie(w, sessionData); err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Path:    r.URL.Path,
			Message: "failed to create session",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	// Return user profile (without password)
	response := LoginResponse{
		User:   userWithPassword.User,
		CartID: cartID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleLogout processes logout requests
func (g *Gateway) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"logged out successfully"}`))
}

// getUserByUsername fetches user details from the user service by username
func (g *Gateway) getUserByUsername(username string) (*UserModificationRequest, error) {
	// Get all users and find by username
	resp, err := http.Get(g.userServiceURL + "/api/v1/core/users")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch users")
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username != nil && *user.Username == username {
			// For this demo, we'll create a UserModificationRequest with user-specific passwords
			// In production, you'd fetch this from a secure user store with the user record
			var password string
			switch *user.Username {
			case "root":
				password = "root"
			case "user":
				password = "user"
			default:
				password = "password123" // Default for other users
			}

			userWithPassword := &UserModificationRequest{
				User:     user,
				Password: &password,
			}
			return userWithPassword, nil
		}
	}

	return nil, errors.New("user not found")
} // validatePassword validates the provided password against the stored hash
func (g *Gateway) validatePassword(password string, hashedPassword *string) bool {
	if hashedPassword == nil {
		return false
	}
	// Simple password validation (in production, use proper bcrypt or similar)
	hash := sha256.Sum256([]byte(password))
	expectedHash := fmt.Sprintf("%x", hash)
	return expectedHash == *hashedPassword || *hashedPassword == password // Allow plain text for demo
}

// getOrCreateCartForUser creates or retrieves a cart for the user
func (g *Gateway) getOrCreateCartForUser(userID string) (string, error) {
	// Try to get existing cart for user
	resp, err := http.Get(g.cartServiceURL + "/api/v1/core/carts")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var carts []Cart
		if err := json.NewDecoder(resp.Body).Decode(&carts); err == nil {
			// Find cart for this user (using OwnerID field)
			for _, cart := range carts {
				if cart.OwnerID.String() == userID {
					return cart.ID.String(), nil
				}
			}
		}
	}

	// Create new cart for user
	cartData := map[string]interface{}{
		"owner_id": userID,
		"items":    []interface{}{},
	}

	cartJSON, _ := json.Marshal(cartData)
	resp, err = http.Post(g.cartServiceURL+"/api/v1/core/carts", "application/json", bytes.NewBuffer(cartJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("failed to create cart")
	}

	var cart Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return "", err
	}

	return cart.ID.String(), nil
}

// setSessionCookie creates and sets a secure session cookie
func (g *Gateway) setSessionCookie(w http.ResponseWriter, sessionData SessionData) error {
	// Encode session data
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	// Simple encoding (in production, use proper encryption/signing)
	encoded := base64.URLEncoding.EncodeToString(sessionJSON)

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

// getSessionData extracts session data from cookie
func (g *Gateway) getSessionData(r *http.Request) (*SessionData, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	// Decode session data
	sessionJSON, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}

	var sessionData SessionData
	if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > sessionData.Exp {
		return nil, errors.New("session expired")
	}

	return &sessionData, nil
}

// proxyToService handles proxying requests to appropriate microservices
func (g *Gateway) proxyToService(w http.ResponseWriter, r *http.Request) {
	var targetURL string

	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/core/users"):
		targetURL = g.userServiceURL
	case strings.HasPrefix(r.URL.Path, "/api/v1/core/carts"):
		targetURL = g.cartServiceURL
	case strings.HasPrefix(r.URL.Path, "/api/v1/core/items"):
		targetURL = g.itemServiceURL
	case strings.HasPrefix(r.URL.Path, "/api/v1/core/checkouts"):
		targetURL = g.checkoutServiceURL
	case strings.HasPrefix(r.URL.Path, "/api/v1/presentation/cart"):
		targetURL = g.cartPresentationServiceURL
	default:
		(&router.ErrorResponse{
			Status:  http.StatusNotFound,
			Path:    r.URL.Path,
			Message: "service not found",
		}).WriteTo(w)
		return
	}

	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		(&router.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Path:    r.URL.Path,
			Message: "invalid target URL",
			Error:   err.Error(),
		}).WriteTo(w)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Set CORS headers at the gateway level
	g.setCORSHeaders(w, r)

	// Handle OPTIONS requests for CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Modify the request to add authentication context if needed
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Add session data to headers if available
		if sessionData, err := g.getSessionData(r); err == nil {
			if sessionData != nil {
				req.Header.Set("X-User-ID", sessionData.UserID)
				req.Header.Set("X-Cart-ID", sessionData.CartID)
				req.Header.Set("X-User-Username", sessionData.Username)
			}
		}

		// Add a header to indicate the request came through the gateway
		req.Header.Set("X-Via-Gateway", "true")
	}

	// Modify the response to remove duplicate CORS headers from backend services
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove CORS headers from backend services to avoid conflicts
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Max-Age")
		return nil
	}

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// setCORSHeaders sets appropriate CORS headers for the gateway
func (g *Gateway) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8088")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Max-Age", "86400")
}
