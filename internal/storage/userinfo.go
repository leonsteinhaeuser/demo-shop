package storage

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// OIDCUser represents a user for OIDC operations
type OIDCUser struct {
	ID                string
	Username          string
	Password          string
	Email             string
	EmailVerified     bool
	PreferredUsername string
	GivenName         string
	FamilyName        string
	Locale            string
	Claims            map[string]interface{}
}

// UserInfoStore implements storage for user information and OIDC operations
type UserInfoStore struct {
	mu    sync.RWMutex
	users map[string]*OIDCUser
}

// NewUserInfoStore creates a new user info store
func NewUserInfoStore() *UserInfoStore {
	store := &UserInfoStore{
		users: make(map[string]*OIDCUser),
	}

	// Add demo users
	demoUsers := []*OIDCUser{
		{
			ID:                "user1",
			Username:          "demo@example.com",
			Password:          "password123",
			Email:             "demo@example.com",
			EmailVerified:     true,
			PreferredUsername: "demo",
			GivenName:         "Demo",
			FamilyName:        "User",
			Locale:            "en",
			Claims: map[string]interface{}{
				"role": "user",
			},
		},
		{
			ID:                "admin1",
			Username:          "admin@example.com",
			Password:          "admin123",
			Email:             "admin@example.com",
			EmailVerified:     true,
			PreferredUsername: "admin",
			GivenName:         "Admin",
			FamilyName:        "User",
			Locale:            "en",
			Claims: map[string]interface{}{
				"role": "admin",
			},
		},
	}

	for _, user := range demoUsers {
		store.users[user.ID] = user
	}

	return store
}

// GetUserByID gets a user by ID
func (s *UserInfoStore) GetUserByID(ctx context.Context, userID string) (*OIDCUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserBySubject gets a user by subject (same as ID in our case)
func (s *UserInfoStore) GetUserBySubject(ctx context.Context, subject string) (*OIDCUser, error) {
	return s.GetUserByID(ctx, subject)
}

// AuthenticateUser authenticates a user by username and password
func (s *UserInfoStore) AuthenticateUser(ctx context.Context, username, password string) (*OIDCUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username && user.Password == password {
			return user, nil
		}
	}
	return nil, errors.New("invalid credentials")
}

// ValidateUser validates user credentials
func (s *UserInfoStore) ValidateUser(ctx context.Context, username, password string) (string, error) {
	user, err := s.AuthenticateUser(ctx, username, password)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

// CreateUser adds a new user
func (s *UserInfoStore) CreateUser(ctx context.Context, user *OIDCUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	if _, exists := s.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	s.users[user.ID] = user
	return nil
}

// UpdateUser updates an existing user
func (s *UserInfoStore) UpdateUser(ctx context.Context, user *OIDCUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	s.users[user.ID] = user
	return nil
}

// DeleteUser removes a user
func (s *UserInfoStore) DeleteUser(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[userID]; !exists {
		return errors.New("user not found")
	}

	delete(s.users, userID)
	return nil
}

// ListUsers returns all users
func (s *UserInfoStore) ListUsers(ctx context.Context) ([]*OIDCUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*OIDCUser, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

// GetClaims returns the user's claims for userinfo endpoint
func (u *OIDCUser) GetClaims() map[string]interface{} {
	claims := make(map[string]interface{})

	// Standard claims
	claims["sub"] = u.ID
	claims["email"] = u.Email
	claims["email_verified"] = u.EmailVerified
	claims["preferred_username"] = u.PreferredUsername
	claims["given_name"] = u.GivenName
	claims["family_name"] = u.FamilyName
	claims["name"] = u.GetName()
	claims["locale"] = u.Locale

	// Custom claims
	for key, value := range u.Claims {
		claims[key] = value
	}

	return claims
}

// GetName returns the full name
func (u *OIDCUser) GetName() string {
	name := strings.TrimSpace(u.GivenName + " " + u.FamilyName)
	if name == "" {
		return u.PreferredUsername
	}
	return name
}
