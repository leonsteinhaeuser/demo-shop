package v1

import (
	"fmt"
	"net/http"

	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/leonsteinhaeuser/demo-shop/internal/storage"
	"github.com/zitadel/oidc/v3/pkg/op"
)

var _ router.ApiObject = &OIDCRouter{}

// OIDCConfig represents OIDC configuration
type OIDCConfig struct {
	Issuer        string `json:"issuer"`
	Port          int    `json:"port"`
	AllowInsecure bool   `json:"allow_insecure"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	AuthRequestID string `json:"authRequestID"`
}

// OIDCRouter implements the OIDC API router
type OIDCRouter struct {
	Storage  *storage.OIDCStorage
	Provider op.OpenIDProvider
	Config   *OIDCConfig
}

// NewOIDCRouter creates a new OIDC router
func NewOIDCRouter(config *OIDCConfig) (*OIDCRouter, error) {
	if config == nil {
		config = &OIDCConfig{
			Issuer:        "http://localhost:8080",
			Port:          8080,
			AllowInsecure: true,
		}
	}

	storage := storage.NewOIDCStorage()

	// Create the OIDC provider configuration
	opConfig := &op.Config{
		CryptoKey: [32]byte{}, // This should be a proper key in production

		DefaultLogoutRedirectURI: config.Issuer + "/logout/callback",
		CodeMethodS256:           true,
		AuthMethodPost:           true,
		AuthMethodPrivateKeyJWT:  true,
		GrantTypeRefreshToken:    true,
		RequestObjectSupported:   true,
	}

	// Create provider options
	options := []op.Option{}
	if config.AllowInsecure {
		options = append(options, op.WithAllowInsecure())
	}

	// Create the OpenID Provider
	provider, err := op.NewOpenIDProvider(config.Issuer, opConfig, storage, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	return &OIDCRouter{
		Storage:  storage,
		Provider: provider,
		Config:   config,
	}, nil
}

func (o *OIDCRouter) GetApiVersion() string {
	return "v1"
}

func (o *OIDCRouter) GetGroup() string {
	return "auth"
}

func (o *OIDCRouter) GetKind() string {
	return "oidc"
}

func (o *OIDCRouter) Routes() []router.PathObject {
	return []router.PathObject{
		// Custom login endpoints
		{
			Path:   "/login",
			Method: "GET",
			Func:   o.loginPage,
		},
		{
			Path:   "/login",
			Method: "POST",
			Func:   o.handleLogin,
		},
		{
			Path:   "/callback",
			Method: "GET",
			Func:   o.loginCallback,
		},
		// OIDC discovery endpoint
		{
			Path:   "/.well-known/openid_configuration",
			Method: "GET",
			Func:   o.discovery,
		},
		// OIDC standard endpoints
		{
			Path:   "/auth",
			Method: "GET",
			Func:   o.authorization,
		},
		{
			Path:   "/auth",
			Method: "POST",
			Func:   o.authorization,
		},
		{
			Path:   "/token",
			Method: "POST",
			Func:   o.token,
		},
		{
			Path:   "/keys",
			Method: "GET",
			Func:   o.keys,
		},
		{
			Path:   "/userinfo",
			Method: "GET",
			Func:   o.userinfo,
		},
		{
			Path:   "/userinfo",
			Method: "POST",
			Func:   o.userinfo,
		},
		{
			Path:   "/revoke",
			Method: "POST",
			Func:   o.revocation,
		},
		{
			Path:   "/introspect",
			Method: "POST",
			Func:   o.introspection,
		},
		{
			Path:   "/end_session",
			Method: "GET",
			Func:   o.endSession,
		},
		{
			Path:   "/end_session",
			Method: "POST",
			Func:   o.endSession,
		},
	}
}

// Login page (simplified HTML form)
func (o *OIDCRouter) loginPage(w http.ResponseWriter, r *http.Request) {
	authRequestID := r.URL.Query().Get("authRequestID")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Login - Demo Shop OIDC</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; }
        input[type="text"], input[type="password"] { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { background-color: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; width: 100%; }
        button:hover { background-color: #0056b3; }
        .error { color: red; margin-top: 10px; }
        .demo-creds { margin-top: 20px; font-size: 14px; color: #666; background-color: #f8f9fa; padding: 15px; border-radius: 4px; }
    </style>
</head>
<body>
    <h2>Login to Demo Shop</h2>
    <form method="post" action="/api/v1/auth/oidc/login">
        <input type="hidden" name="authRequestID" value="` + authRequestID + `">
        <div class="form-group">
            <label for="username">Username/Email:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div class="form-group">
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Login</button>
    </form>
    <div class="demo-creds">
        <strong>Demo credentials:</strong><br>
        <strong>User:</strong> demo@example.com / password123<br>
        <strong>Admin:</strong> admin@example.com / admin123
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Handle login form submission
func (o *OIDCRouter) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	authRequestID := r.FormValue("authRequestID")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Validate user credentials
	userID, err := o.Storage.ValidateUser(r.Context(), username, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if authRequestID != "" {
		// Redirect to callback with user ID
		callbackURL := fmt.Sprintf("/api/v1/auth/oidc/callback?authRequestID=%s&userID=%s", authRequestID, userID)
		http.Redirect(w, r, callbackURL, http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

// Login callback handler
func (o *OIDCRouter) loginCallback(w http.ResponseWriter, r *http.Request) {
	authRequestID := r.URL.Query().Get("authRequestID")
	userID := r.URL.Query().Get("userID")

	if authRequestID == "" || userID == "" {
		http.Error(w, "Missing authRequestID or userID", http.StatusBadRequest)
		return
	}

	// Get the auth request
	authReq, err := o.Storage.AuthRequestByID(r.Context(), authRequestID)
	if err != nil {
		http.Error(w, "Invalid auth request", http.StatusBadRequest)
		return
	}

	// Set the user ID and mark as done
	if ar, ok := authReq.(*storage.AuthRequest); ok {
		ar.UserID = userID
		ar.IsDone = true
	}

	// Generate authorization response
	op.AuthorizeCallback(w, r, o.Provider)
}

// OIDC endpoints that delegate to the provider
func (o *OIDCRouter) discovery(w http.ResponseWriter, r *http.Request) {
	config := op.CreateDiscoveryConfig(r.Context(), o.Provider, o.Storage)
	op.Discover(w, config)
}

func (o *OIDCRouter) authorization(w http.ResponseWriter, r *http.Request) {
	op.Authorize(w, r, o.Provider)
}

func (o *OIDCRouter) token(w http.ResponseWriter, r *http.Request) {
	op.Exchange(w, r, o.Provider)
}

func (o *OIDCRouter) keys(w http.ResponseWriter, r *http.Request) {
	op.Keys(w, r, o.Storage)
}

func (o *OIDCRouter) userinfo(w http.ResponseWriter, r *http.Request) {
	op.Userinfo(w, r, o.Provider)
}

func (o *OIDCRouter) revocation(w http.ResponseWriter, r *http.Request) {
	op.Revoke(w, r, o.Provider)
}

func (o *OIDCRouter) introspection(w http.ResponseWriter, r *http.Request) {
	op.Introspect(w, r, o.Provider)
}

func (o *OIDCRouter) endSession(w http.ResponseWriter, r *http.Request) {
	op.EndSession(w, r, o.Provider)
}
