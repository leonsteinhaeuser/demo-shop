package storage

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/text/language"
)

// Token represents an access or refresh token
type Token struct {
	ID           string
	UserID       string
	ClientID     string
	Scopes       []string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	TokenType    string // "access" or "refresh"
	RefreshToken string // Only set for access tokens that have a corresponding refresh token
}

// OIDCStorage implements all required OIDC storage interfaces
type OIDCStorage struct {
	clientStore      *ClientStore
	authRequestStore *AuthRequestStore
	userInfoStore    *UserInfoStore

	// Token storage
	tokensMu      sync.RWMutex
	tokens        map[string]*Token // tokenID -> Token
	refreshTokens map[string]*Token // refreshToken -> Token

	// Keys for signing
	keysMu     sync.RWMutex
	signingKey *rsa.PrivateKey
	keyID      string
}

// NewOIDCStorage creates a new OIDC storage
func NewOIDCStorage() *OIDCStorage {
	// Generate RSA key for signing
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate RSA key: %v", err))
	}

	return &OIDCStorage{
		clientStore:      NewClientStore(),
		authRequestStore: NewAuthRequestStore(),
		userInfoStore:    NewUserInfoStore(),
		tokens:           make(map[string]*Token),
		refreshTokens:    make(map[string]*Token),
		signingKey:       key,
		keyID:            uuid.New().String(),
	}
}

// Client storage methods
func (s *OIDCStorage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	return s.clientStore.GetClientByClientID(ctx, clientID)
}

func (s *OIDCStorage) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
	return s.clientStore.AuthorizeClientIDSecret(ctx, clientID, clientSecret)
}

// Auth request storage methods
func (s *OIDCStorage) CreateAuthRequest(ctx context.Context, authReq *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	// Convert oidc.AuthRequest to our AuthRequest type
	id := uuid.New().String()

	request := &AuthRequest{
		ID:            id,
		CreationDate:  time.Now(),
		ClientID:      authReq.ClientID,
		RedirectURI:   authReq.RedirectURI,
		State:         authReq.State,
		Nonce:         authReq.Nonce,
		Scopes:        authReq.Scopes,
		ResponseType:  authReq.ResponseType,
		ResponseMode:  authReq.ResponseMode,
		CodeChallenge: nil, // Will be set if PKCE is used
		UserID:        userID,
		LoginHint:     "",
		IsDone:        false,
	}

	s.authRequestStore.mu.Lock()
	s.authRequestStore.requests[id] = request
	s.authRequestStore.mu.Unlock()

	return request, nil
}

func (s *OIDCStorage) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
	return s.authRequestStore.AuthRequestByID(ctx, id)
}

func (s *OIDCStorage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	return s.authRequestStore.AuthRequestByCode(ctx, code)
}

func (s *OIDCStorage) SaveAuthCode(ctx context.Context, id string, code string) error {
	return s.authRequestStore.SaveAuthCode(ctx, id, code)
}

func (s *OIDCStorage) DeleteAuthRequest(ctx context.Context, id string) error {
	return s.authRequestStore.DeleteAuthRequest(ctx, id)
}

// Token storage methods
func (s *OIDCStorage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	tokenID := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour) // 1 hour expiration

	// Extract user and client information from the request
	var userID, clientID string
	var scopes []string

	switch req := request.(type) {
	case *AuthRequest:
		userID = req.UserID
		clientID = req.ClientID
		scopes = req.Scopes
	default:
		return "", time.Time{}, errors.New("unsupported token request type")
	}

	token := &Token{
		ID:        tokenID,
		UserID:    userID,
		ClientID:  clientID,
		Scopes:    scopes,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		TokenType: "access",
	}

	s.tokens[tokenID] = token
	return tokenID, expiresAt, nil
}

func (s *OIDCStorage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshTokenID string, expiration time.Time, err error) {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	accessTokenID = uuid.New().String()
	refreshTokenID := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour) // 1 hour expiration

	// Extract user and client information from the request
	var userID, clientID string
	var scopes []string

	switch req := request.(type) {
	case *AuthRequest:
		userID = req.UserID
		clientID = req.ClientID
		scopes = req.Scopes
	default:
		return "", "", time.Time{}, errors.New("unsupported token request type")
	}

	// Create access token
	accessToken := &Token{
		ID:           accessTokenID,
		UserID:       userID,
		ClientID:     clientID,
		Scopes:       scopes,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		TokenType:    "access",
		RefreshToken: refreshTokenID,
	}

	// Create refresh token (longer expiration)
	refreshToken := &Token{
		ID:        refreshTokenID,
		UserID:    userID,
		ClientID:  clientID,
		Scopes:    scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour * 30), // 30 days
		TokenType: "refresh",
	}

	s.tokens[accessTokenID] = accessToken
	s.tokens[refreshTokenID] = refreshToken
	s.refreshTokens[refreshTokenID] = refreshToken

	// Remove old refresh token if provided
	if currentRefreshToken != "" {
		if oldToken, exists := s.refreshTokens[currentRefreshToken]; exists {
			delete(s.tokens, oldToken.ID)
			delete(s.refreshTokens, currentRefreshToken)
		}
	}

	return accessTokenID, refreshTokenID, expiresAt, nil
}

func (s *OIDCStorage) TokenRequestByRefreshToken(ctx context.Context, refreshTokenID string) (op.RefreshTokenRequest, error) {
	s.tokensMu.RLock()
	defer s.tokensMu.RUnlock()

	token, exists := s.refreshTokens[refreshTokenID]
	if !exists {
		return nil, errors.New("refresh token not found")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	// Return a simple refresh token request implementation
	return &RefreshTokenRequest{
		RefreshToken: refreshTokenID,
		UserID:       token.UserID,
		ClientID:     token.ClientID,
		Scopes:       token.Scopes,
	}, nil
}

func (s *OIDCStorage) TerminateSession(ctx context.Context, userID string, clientID string) error {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	// Remove all tokens for the user and client
	for id, token := range s.tokens {
		if token.UserID == userID && token.ClientID == clientID {
			delete(s.tokens, id)
			if token.TokenType == "refresh" {
				delete(s.refreshTokens, token.ID)
			}
		}
	}

	return nil
}

func (s *OIDCStorage) RevokeToken(ctx context.Context, tokenOrTokenID string, userID string, clientID string) *oidc.Error {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	// Try to find token by ID first
	if token, exists := s.tokens[tokenOrTokenID]; exists {
		delete(s.tokens, tokenOrTokenID)
		if token.TokenType == "refresh" {
			delete(s.refreshTokens, tokenOrTokenID)
		}
		return nil
	}

	// Try to find refresh token by token value
	if token, exists := s.refreshTokens[tokenOrTokenID]; exists {
		delete(s.tokens, token.ID)
		delete(s.refreshTokens, tokenOrTokenID)
		return nil
	}

	return &oidc.Error{
		ErrorType:   oidc.InvalidRequest,
		Description: "token not found",
	}
}

func (s *OIDCStorage) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (userID string, tokenID string, err error) {
	s.tokensMu.RLock()
	defer s.tokensMu.RUnlock()

	refreshToken, exists := s.refreshTokens[token]
	if !exists {
		return "", "", op.ErrInvalidRefreshToken
	}

	if refreshToken.ClientID != clientID {
		return "", "", op.ErrInvalidRefreshToken
	}

	return refreshToken.UserID, refreshToken.ID, nil
}

// Signing key methods
func (s *OIDCStorage) SigningKey(ctx context.Context) (op.SigningKey, error) {
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()

	return &SigningKey{
		key:   s.signingKey,
		keyID: s.keyID,
		alg:   jose.RS256,
	}, nil
}

func (s *OIDCStorage) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.RS256}, nil
}

func (s *OIDCStorage) KeySet(ctx context.Context) ([]op.Key, error) {
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()

	publicKey := &s.signingKey.PublicKey

	key := &Key{
		keyID: s.keyID,
		alg:   jose.RS256,
		key:   publicKey,
	}

	return []op.Key{key}, nil
}

// User methods for userinfo endpoint
func (s *OIDCStorage) GetUserBySubject(ctx context.Context, subject string) (*OIDCUser, error) {
	return s.userInfoStore.GetUserBySubject(ctx, subject)
}

func (s *OIDCStorage) ValidateUser(ctx context.Context, username, password string) (string, error) {
	return s.userInfoStore.ValidateUser(ctx, username, password)
}

// OPStorage interface methods
func (s *OIDCStorage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	// Deprecated method - empty implementation
	return nil
}

func (s *OIDCStorage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, subject, origin string) error {
	user, err := s.userInfoStore.GetUserByID(ctx, subject)
	if err != nil {
		return err
	}

	claims := user.GetClaims()

	// Set standard claims
	if sub, ok := claims["sub"].(string); ok {
		userinfo.Subject = sub
	}
	if email, ok := claims["email"].(string); ok {
		userinfo.Email = email
	}
	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userinfo.EmailVerified = oidc.Bool(emailVerified)
	}
	if name, ok := claims["name"].(string); ok {
		userinfo.Name = name
	}
	if givenName, ok := claims["given_name"].(string); ok {
		userinfo.GivenName = givenName
	}
	if familyName, ok := claims["family_name"].(string); ok {
		userinfo.FamilyName = familyName
	}
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		userinfo.PreferredUsername = preferredUsername
	}
	if locale, ok := claims["locale"].(string); ok {
		// Parse locale string to language.Tag
		if tag, err := language.Parse(locale); err == nil {
			userinfo.Locale = oidc.NewLocale(tag)
		}
	}

	return nil
}

func (s *OIDCStorage) SetIntrospectionFromToken(ctx context.Context, introspectionResponse *oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
	s.tokensMu.RLock()
	defer s.tokensMu.RUnlock()

	token, exists := s.tokens[tokenID]
	if !exists {
		introspectionResponse.Active = false
		return nil
	}

	if time.Now().After(token.ExpiresAt) {
		introspectionResponse.Active = false
		return nil
	}

	introspectionResponse.Active = true
	introspectionResponse.ClientID = token.ClientID
	introspectionResponse.Subject = token.UserID
	introspectionResponse.Expiration = oidc.FromTime(token.ExpiresAt)
	introspectionResponse.IssuedAt = oidc.FromTime(token.CreatedAt)
	introspectionResponse.TokenType = oidc.BearerToken
	introspectionResponse.Scope = token.Scopes

	return nil
}

func (s *OIDCStorage) GetPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (map[string]any, error) {
	user, err := s.userInfoStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	claims := make(map[string]any)
	userClaims := user.GetClaims()

	// Return custom claims based on scopes
	for _, scope := range scopes {
		switch scope {
		case "profile":
			if name, ok := userClaims["name"]; ok {
				claims["name"] = name
			}
			if givenName, ok := userClaims["given_name"]; ok {
				claims["given_name"] = givenName
			}
			if familyName, ok := userClaims["family_name"]; ok {
				claims["family_name"] = familyName
			}
			if preferredUsername, ok := userClaims["preferred_username"]; ok {
				claims["preferred_username"] = preferredUsername
			}
			if locale, ok := userClaims["locale"]; ok {
				claims["locale"] = locale
			}
		case "email":
			if email, ok := userClaims["email"]; ok {
				claims["email"] = email
			}
			if emailVerified, ok := userClaims["email_verified"]; ok {
				claims["email_verified"] = emailVerified
			}
		}
	}

	// Add custom claims
	for key, value := range user.Claims {
		claims[key] = value
	}

	return claims, nil
}

func (s *OIDCStorage) GetKeyByIDAndClientID(ctx context.Context, keyID, clientID string) (*jose.JSONWebKey, error) {
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()

	if keyID != s.keyID {
		return nil, errors.New("key not found")
	}

	jwk := &jose.JSONWebKey{
		KeyID:     s.keyID,
		Algorithm: string(jose.RS256),
		Use:       "sig",
		Key:       &s.signingKey.PublicKey,
	}

	return jwk, nil
}

func (s *OIDCStorage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	// Simple validation - return all requested scopes if user exists
	_, err := s.userInfoStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return scopes, nil
}

// Health check method
func (s *OIDCStorage) Health(ctx context.Context) error {
	return nil
}

// RefreshTokenRequest implementation
type RefreshTokenRequest struct {
	RefreshToken string
	UserID       string
	ClientID     string
	Scopes       []string
}

func (r *RefreshTokenRequest) GetAMR() []string                 { return []string{} }
func (r *RefreshTokenRequest) GetAudience() []string            { return []string{r.ClientID} }
func (r *RefreshTokenRequest) GetAuthTime() time.Time           { return time.Now() }
func (r *RefreshTokenRequest) GetClientID() string              { return r.ClientID }
func (r *RefreshTokenRequest) GetScopes() []string              { return r.Scopes }
func (r *RefreshTokenRequest) GetSubject() string               { return r.UserID }
func (r *RefreshTokenRequest) SetCurrentScopes(scopes []string) { r.Scopes = scopes }

// SigningKey implementation
type SigningKey struct {
	key   *rsa.PrivateKey
	keyID string
	alg   jose.SignatureAlgorithm
}

func (s *SigningKey) SignatureAlgorithm() jose.SignatureAlgorithm { return s.alg }
func (s *SigningKey) Key() interface{}                            { return s.key }
func (s *SigningKey) ID() string                                  { return s.keyID }

// Key implementation
type Key struct {
	keyID string
	alg   jose.SignatureAlgorithm
	key   interface{}
}

func (k *Key) Algorithm() jose.SignatureAlgorithm { return k.alg }
func (k *Key) Use() string                        { return "sig" }
func (k *Key) Key() interface{}                   { return k.key }
func (k *Key) ID() string                         { return k.keyID }
