package storage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// Client represents an OIDC client
type Client struct {
	ClientID                                  string                            `json:"id"`
	ClientSecret                              string                            `json:"secret"`
	ClientRedirectURIs                        []string                          `json:"redirect_uris"`
	ClientApplicationType                     op.ApplicationType                `json:"application_type"`
	ClientAuthMethod                          oidc.AuthMethod                   `json:"auth_method"`
	ClientResponseTypes                       []oidc.ResponseType               `json:"response_types"`
	ClientGrantTypes                          []oidc.GrantType                  `json:"grant_types"`
	ClientLoginURL                            func(authRequestID string) string `json:"-"`
	ClientAccessTokenType                     op.AccessTokenType                `json:"access_token_type"`
	ClientIDTokenUserinfoClaimsAssertion      bool                              `json:"id_token_userinfo_claims_assertion"`
	ClientDevMode                             bool                              `json:"dev_mode"`
	ClientRestrictAdditionalIdTokenScopes     func(scopes []string) []string    `json:"-"`
	ClientRestrictAdditionalAccessTokenScopes func(scopes []string) []string    `json:"-"`
	ClientIsScopeAllowed                      func(scope string) bool           `json:"-"`
	ClientIDTokenLifetime                     time.Duration                     `json:"id_token_lifetime"`
	ClientClockSkew                           time.Duration                     `json:"clock_skew"`
}

// ClientStore implements storage for OIDC clients
type ClientStore struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewClientStore creates a new client store
func NewClientStore() *ClientStore {
	store := &ClientStore{
		clients: make(map[string]*Client),
	}

	// Add a default client for demo purposes
	defaultClient := &Client{
		ClientID:                             "demo-client",
		ClientSecret:                         "demo-secret",
		ClientRedirectURIs:                   []string{"http://localhost:8080/callback", "http://localhost:3000/callback"},
		ClientApplicationType:                op.ApplicationTypeWeb,
		ClientAuthMethod:                     oidc.AuthMethodBasic,
		ClientResponseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode},
		ClientGrantTypes:                     []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		ClientAccessTokenType:                op.AccessTokenTypeBearer,
		ClientIDTokenUserinfoClaimsAssertion: false,
		ClientDevMode:                        true,
		ClientIsScopeAllowed: func(scope string) bool {
			return scope == oidc.ScopeOpenID || scope == oidc.ScopeProfile || scope == oidc.ScopeEmail
		},
		ClientIDTokenLifetime: time.Hour,
		ClientClockSkew:       time.Minute,
	}
	store.clients[defaultClient.ClientID] = defaultClient

	return store
}

// GetClientByClientID implements op.Storage interface
func (s *ClientStore) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, errors.New("client not found")
	}
	return client, nil
}

// AuthorizeClientIDSecret implements op.Storage interface
func (s *ClientStore) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientID]
	if !exists {
		return errors.New("client not found")
	}
	if client.ClientSecret != clientSecret {
		return errors.New("invalid client secret")
	}
	return nil
}

// CreateClient adds a new client to the store
func (s *ClientStore) CreateClient(ctx context.Context, client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client.ClientID == "" {
		client.ClientID = uuid.New().String()
	}

	s.clients[client.ClientID] = client
	return nil
}

// UpdateClient updates an existing client
func (s *ClientStore) UpdateClient(ctx context.Context, client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.ClientID]; !exists {
		return errors.New("client not found")
	}

	s.clients[client.ClientID] = client
	return nil
}

// DeleteClient removes a client from the store
func (s *ClientStore) DeleteClient(ctx context.Context, clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[clientID]; !exists {
		return errors.New("client not found")
	}

	delete(s.clients, clientID)
	return nil
}

// ListClients returns all clients
func (s *ClientStore) ListClients(ctx context.Context) ([]*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients, nil
}

// Client interface implementation
func (c *Client) GetID() string {
	return c.ClientID
}

func (c *Client) RedirectURIs() []string {
	return c.ClientRedirectURIs
}

func (c *Client) PostLogoutRedirectURIs() []string {
	return []string{}
}

func (c *Client) ApplicationType() op.ApplicationType {
	return c.ClientApplicationType
}

func (c *Client) AuthMethod() oidc.AuthMethod {
	return c.ClientAuthMethod
}

func (c *Client) ResponseTypes() []oidc.ResponseType {
	return c.ClientResponseTypes
}

func (c *Client) GrantTypes() []oidc.GrantType {
	return c.ClientGrantTypes
}

func (c *Client) LoginURL(id string) string {
	if c.ClientLoginURL != nil {
		return c.ClientLoginURL(id)
	}
	return "/login?authRequestID=" + id
}

func (c *Client) AccessTokenType() op.AccessTokenType {
	return c.ClientAccessTokenType
}

func (c *Client) IDTokenUserinfoClaimsAssertion() bool {
	return c.ClientIDTokenUserinfoClaimsAssertion
}

func (c *Client) DevMode() bool {
	return c.ClientDevMode
}

func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return c.ClientRestrictAdditionalIdTokenScopes
}

func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return c.ClientRestrictAdditionalAccessTokenScopes
}

func (c *Client) IsScopeAllowed(scope string) bool {
	if c.ClientIsScopeAllowed != nil {
		return c.ClientIsScopeAllowed(scope)
	}
	// Default: allow standard OIDC scopes
	return scope == oidc.ScopeOpenID || scope == oidc.ScopeProfile || scope == oidc.ScopeEmail
}

func (c *Client) IDTokenLifetime() time.Duration {
	return c.ClientIDTokenLifetime
}

func (c *Client) ClockSkew() time.Duration {
	return c.ClientClockSkew
}
