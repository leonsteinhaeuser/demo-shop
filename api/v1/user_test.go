package v1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
)

// Test constants
const (
	testPassword = "securepassword123"
	testUsername = "testuser"
	testEmail    = "test@example.com"
)

// MockUserStore implements UserStore interface for testing
type MockUserStore struct {
	users  map[uuid.UUID]*User
	fail   bool
	failOn string
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
		users: make(map[uuid.UUID]*User),
	}
}

func (m *MockUserStore) SetFailure(failOn string) {
	m.fail = true
	m.failOn = failOn
}

func (m *MockUserStore) Create(ctx context.Context, user *UserModificationRequest) error {
	if m.fail && m.failOn == "create" {
		return errors.New("mock create error")
	}
	userObj := &User{
		ID:            user.ID,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		Username:      user.Username,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		PreferredName: user.PreferredName,
		GivenName:     user.GivenName,
		FamilyName:    user.FamilyName,
		Locale:        user.Locale,
		IsAdmin:       user.IsAdmin,
	}
	m.users[user.ID] = userObj
	return nil
}

func (m *MockUserStore) List(ctx context.Context, page, limit int) ([]User, error) {
	if m.fail && m.failOn == "list" {
		return nil, errors.New("mock list error")
	}
	users := make([]User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *MockUserStore) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	if m.fail && m.failOn == "get" {
		return nil, errors.New("mock get error")
	}
	user, exists := m.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserStore) Update(ctx context.Context, user *UserModificationRequest) error {
	if m.fail && m.failOn == "update" {
		return errors.New("mock update error")
	}
	if existingUser, exists := m.users[user.ID]; exists {
		// Update only non-nil fields
		if user.Username != nil {
			existingUser.Username = user.Username
		}
		if user.Email != nil {
			existingUser.Email = user.Email
		}
		existingUser.UpdatedAt = user.UpdatedAt
	}
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.fail && m.failOn == "delete" {
		return errors.New("mock delete error")
	}
	delete(m.users, id)
	return nil
}

func TestNewUserRouter(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	if router == nil {
		t.Fatal("Expected router to be created")
	}

	if router.UserStore != store {
		t.Error("Expected store to be set correctly")
	}
}

func TestUserRouter_GetApiVersion(t *testing.T) {
	router := NewUserRouter(NewMockUserStore())
	if router.GetApiVersion() != "v1" {
		t.Errorf("Expected API version v1, got %s", router.GetApiVersion())
	}
}

func TestUserRouter_GetGroup(t *testing.T) {
	router := NewUserRouter(NewMockUserStore())
	if router.GetGroup() != "core" {
		t.Errorf("Expected group core, got %s", router.GetGroup())
	}
}

func TestUserRouter_GetKind(t *testing.T) {
	router := NewUserRouter(NewMockUserStore())
	if router.GetKind() != "users" {
		t.Errorf("Expected kind users, got %s", router.GetKind())
	}
}

func TestUserRouter_createUser_Success(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	password := testPassword
	username := testUsername
	email := testEmail

	user := &UserModificationRequest{
		User: User{
			ID:       uuid.New(),
			Username: &username,
			Email:    &email,
		},
		Password: &password,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/users", nil)
	err := router.createUser(context.Background(), req, user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify user was stored
	storedUser, exists := store.users[user.ID]
	if !exists {
		t.Error("Expected user to be stored")
	}

	if *storedUser.Username != username {
		t.Errorf("Expected username %s, got %s", username, *storedUser.Username)
	}

	if *storedUser.Email != email {
		t.Errorf("Expected email %s, got %s", email, *storedUser.Email)
	}
}

func TestUserRouter_createUser_ShortPassword(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	password := "short" // Too short password
	username := testUsername
	email := testEmail

	user := &UserModificationRequest{
		User: User{
			ID:       uuid.New(),
			Username: &username,
			Email:    &email,
		},
		Password: &password,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/users", nil)
	err := router.createUser(context.Background(), req, user)

	if err == nil {
		t.Error("Expected error for short password")
	}
}

func TestUserRouter_createUser_EmptyUsername(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	password := testPassword
	username := ""
	email := testEmail

	user := &UserModificationRequest{
		User: User{
			ID:       uuid.New(),
			Username: &username,
			Email:    &email,
		},
		Password: &password,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/users", nil)
	err := router.createUser(context.Background(), req, user)

	if err == nil {
		t.Error("Expected error for empty username")
	}
}

func TestUserRouter_createUser_EmptyEmail(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	password := testPassword
	username := testUsername
	email := ""

	user := &UserModificationRequest{
		User: User{
			ID:       uuid.New(),
			Username: &username,
			Email:    &email,
		},
		Password: &password,
	}

	req := httptest.NewRequest("POST", "/api/v1/core/users", nil)
	err := router.createUser(context.Background(), req, user)

	if err == nil {
		t.Error("Expected error for empty email")
	}
}

func TestUserRouter_listUsers_Success(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	// Add some test users
	username1 := "user1"
	email1 := "user1@example.com"
	username2 := "user2"
	email2 := "user2@example.com"

	user1 := &User{ID: uuid.New(), Username: &username1, Email: &email1}
	user2 := &User{ID: uuid.New(), Username: &username2, Email: &email2}
	store.users[user1.ID] = user1
	store.users[user2.ID] = user2

	req := httptest.NewRequest("GET", "/api/v1/core/users", nil)
	filters := handlers.FilterObjectList{Page: 1, Limit: 10}

	users, err := router.listUsers(context.Background(), req, filters)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestUserRouter_getUser_Success(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	userID := uuid.New()
	username := testUsername
	email := testEmail
	expectedUser := &User{
		ID:        userID,
		Username:  &username,
		Email:     &email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.users[userID] = expectedUser

	req := httptest.NewRequest("GET", "/api/v1/core/users/"+userID.String(), nil)
	req.SetPathValue("id", userID.String())

	user, err := router.getUser(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.ID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, user.ID)
	}

	if *user.Username != username {
		t.Errorf("Expected username %s, got %s", username, *user.Username)
	}
}

func TestUserRouter_getUser_NotFound(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	userID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/core/users/"+userID.String(), nil)
	req.SetPathValue("id", userID.String())

	user, err := router.getUser(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error for user not found, got %v", err)
	}

	if user != nil {
		t.Error("Expected nil user for not found")
	}
}

func TestUserRouter_updateUser_Success(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	userID := uuid.New()
	originalUsername := "original"
	originalEmail := "original@example.com"
	originalUser := &User{
		ID:        userID,
		Username:  &originalUsername,
		Email:     &originalEmail,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.users[userID] = originalUser

	newUsername := "updated"
	newEmail := "updated@example.com"
	updatedUser := &UserModificationRequest{
		User: User{
			ID:       userID,
			Username: &newUsername,
			Email:    &newEmail,
		},
	}

	req := httptest.NewRequest("PUT", "/api/v1/core/users/"+userID.String(), nil)
	req.SetPathValue("id", userID.String())

	err := router.updateUser(context.Background(), req, updatedUser)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify user was updated
	storedUser := store.users[userID]
	if *storedUser.Username != newUsername {
		t.Errorf("Expected updated username %s, got %s", newUsername, *storedUser.Username)
	}
}

func TestUserRouter_deleteUser_Success(t *testing.T) {
	store := NewMockUserStore()
	router := NewUserRouter(store)

	userID := uuid.New()
	username := testUsername
	email := testEmail
	user := &User{
		ID:       userID,
		Username: &username,
		Email:    &email,
	}
	store.users[userID] = user

	req := httptest.NewRequest("DELETE", "/api/v1/core/users/"+userID.String(), nil)
	req.SetPathValue("id", userID.String())

	deleteReq := &UserDeleteRequest{ID: userID}
	err := router.deleteUser(context.Background(), req, deleteReq)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify user was deleted
	_, exists := store.users[userID]
	if exists {
		t.Error("Expected user to be deleted")
	}
}
