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

// AuthRequest represents an OIDC authorization request
type AuthRequest struct {
	ID            string
	CreationDate  time.Time
	ClientID      string
	RedirectURI   string
	State         string
	Nonce         string
	Scopes        []string
	ResponseType  oidc.ResponseType
	ResponseMode  oidc.ResponseMode
	CodeChallenge *oidc.CodeChallenge

	// User information when authenticated
	UserID    string
	LoginHint string
	MaxAge    *time.Duration

	// Current state
	IsDone bool
}

// AuthRequestStore implements storage for authorization requests
type AuthRequestStore struct {
	mu       sync.RWMutex
	requests map[string]*AuthRequest
}

// NewAuthRequestStore creates a new auth request store
func NewAuthRequestStore() *AuthRequestStore {
	return &AuthRequestStore{
		requests: make(map[string]*AuthRequest),
	}
}

// CreateAuthRequest implements op.Storage interface
func (s *AuthRequestStore) CreateAuthRequest(ctx context.Context, authReq op.AuthRequest, userID string) (op.AuthRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()

	request := &AuthRequest{
		ID:            id,
		CreationDate:  time.Now(),
		ClientID:      authReq.GetClientID(),
		RedirectURI:   authReq.GetRedirectURI(),
		State:         authReq.GetState(),
		Nonce:         authReq.GetNonce(),
		Scopes:        authReq.GetScopes(),
		ResponseType:  authReq.GetResponseType(),
		ResponseMode:  authReq.GetResponseMode(),
		CodeChallenge: authReq.GetCodeChallenge(),
		UserID:        userID,
		LoginHint:     authReq.GetSubject(),
		IsDone:        false,
	}

	s.requests[id] = request
	return request, nil
}

// AuthRequestByID implements op.Storage interface
func (s *AuthRequestStore) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, exists := s.requests[id]
	if !exists {
		return nil, errors.New("auth request not found")
	}
	return request, nil
}

// AuthRequestByCode implements op.Storage interface
func (s *AuthRequestStore) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// In a real implementation, you would store a mapping of code to auth request ID
	// For simplicity, we'll treat the code as the request ID
	request, exists := s.requests[code]
	if !exists {
		return nil, errors.New("auth request not found by code")
	}
	return request, nil
}

// SaveAuthCode implements op.Storage interface
func (s *AuthRequestStore) SaveAuthCode(ctx context.Context, id string, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real implementation, you would save the code-to-id mapping
	// For simplicity, we'll just mark the request as done
	if request, exists := s.requests[id]; exists {
		request.IsDone = true
		return nil
	}
	return errors.New("auth request not found")
}

// DeleteAuthRequest implements op.Storage interface
func (s *AuthRequestStore) DeleteAuthRequest(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.requests[id]; !exists {
		return errors.New("auth request not found")
	}
	delete(s.requests, id)
	return nil
}

// AuthRequest interface implementation
func (r *AuthRequest) GetID() string {
	return r.ID
}

func (r *AuthRequest) GetACR() string {
	return "" // Authentication Context Class Reference
}

func (r *AuthRequest) GetAMR() []string {
	return []string{} // Authentication Methods References
}

func (r *AuthRequest) GetAudience() []string {
	return []string{r.ClientID}
}

func (r *AuthRequest) GetAuthTime() time.Time {
	return r.CreationDate
}

func (r *AuthRequest) GetClientID() string {
	return r.ClientID
}

func (r *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	return r.CodeChallenge
}

func (r *AuthRequest) GetNonce() string {
	return r.Nonce
}

func (r *AuthRequest) GetRedirectURI() string {
	return r.RedirectURI
}

func (r *AuthRequest) GetResponseMode() oidc.ResponseMode {
	return r.ResponseMode
}

func (r *AuthRequest) GetResponseType() oidc.ResponseType {
	return r.ResponseType
}

func (r *AuthRequest) GetScopes() []string {
	return r.Scopes
}

func (r *AuthRequest) GetState() string {
	return r.State
}

func (r *AuthRequest) GetSubject() string {
	return r.UserID
}

func (r *AuthRequest) Done() bool {
	return r.IsDone
}
