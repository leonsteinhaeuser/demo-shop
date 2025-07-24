package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username      *string `json:"username"`
	Email         *string `json:"email"`
	EmailVerified bool    `json:"email_verified"`

	PreferredName *string `json:"preferred_name,omitempty"`
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
	Locale        *string `json:"locale,omitempty"`
}

type UserModificationRequest struct {
	User
	Password *string `json:"password,omitempty"`
}

// UserStore interface for user operations
type UserStore interface {
	Create(ctx context.Context, item *UserModificationRequest) error
	List(ctx context.Context, page, limit int) ([]User, error)
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, item *UserModificationRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRouter implements the API router for user endpoints
type UserRouter struct {
	UserStore UserStore
}

func NewUserRouter(userStore UserStore) *UserRouter {
	return &UserRouter{
		UserStore: userStore,
	}
}

func (u *UserRouter) GetApiVersion() string {
	return "v1"
}

func (u *UserRouter) GetGroup() string {
	return "core"
}

func (u *UserRouter) GetKind() string {
	return "users"
}

func (u *UserRouter) Routes() []router.PathObject {
	return []router.PathObject{
		{
			Method: "POST",
			Func:   handlers.HttpPost(u.createUser),
		},
		{
			Method: "GET",
			Func:   handlers.HttpList(u.listUsers),
		},
		{
			Path:   "/{id}",
			Method: "GET",
			Func:   handlers.HttpGet(u.getUser),
		},
		{
			Path:   "/{id}",
			Method: "PUT",
			Func:   handlers.HttpUpdate(u.updateUser),
		},
		{
			Path:   "/{id}",
			Method: "DELETE",
			Func:   handlers.HttpDelete(u.deleteUser),
		},
	}
}

// UserValidationRequest represents a request to validate user credentials
type UserValidationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *UserRouter) createUser(ctx context.Context, r *http.Request, user *UserModificationRequest) error {
	if u.UserStore == nil {
		return router.ErrObjectStorageNotImplemented
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if user.Password == nil || *user.Password == "" {
		return errors.New("password is required")
	}
	// ensure password meets security requirements (e.g., length, complexity) here if needed
	// For simplicity, let's say it must be at least 12 characters long
	if len(*user.Password) < 12 {
		return errors.New("password must be at least 12 characters long")
	}
	if user.Username == nil || *user.Username == "" {
		return errors.New("username cannot be empty")
	}
	if user.Email == nil || *user.Email == "" {
		return errors.New("email cannot be empty")
	}
	return u.UserStore.Create(ctx, user)
}

func (u *UserRouter) listUsers(ctx context.Context, r *http.Request, filters handlers.FilterObjectList) ([]User, error) {
	if u.UserStore == nil {
		return nil, router.ErrObjectStorageNotImplemented
	}

	return u.UserStore.List(ctx, filters.Page, filters.Limit)
}

func (u *UserRouter) getUser(ctx context.Context, r *http.Request) (*User, error) {
	if u.UserStore == nil {
		return nil, router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return nil, err
	}

	return u.UserStore.Get(ctx, id)
}

func (u *UserRouter) updateUser(ctx context.Context, r *http.Request, user *UserModificationRequest) error {
	if u.UserStore == nil {
		return router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return err
	}
	if user.ID != id {
		return errors.New("ID in path does not match user ID")
	}

	user.UpdatedAt = time.Now()

	// if the password is not provided, we do not update it
	if user.Password != nil {
		if *user.Password == "" {
			return errors.New("password is required")
		}
		if len(*user.Password) < 12 {
			return errors.New("password must be at least 12 characters long")
		}

	}

	if user.Username != nil && *user.Username == "" {
		return errors.New("username cannot be empty")
	}
	if user.Email != nil && *user.Email == "" {
		return errors.New("email cannot be empty")
	}
	if user.PreferredName != nil && *user.PreferredName == "" {
		return errors.New("preferred_name cannot be empty")
	}
	if user.GivenName != nil && *user.GivenName == "" {
		return errors.New("given_name cannot be empty")
	}
	if user.FamilyName != nil && *user.FamilyName == "" {
		return errors.New("family_name cannot be empty")
	}
	if user.Locale != nil && *user.Locale == "" {
		return errors.New("locale cannot be empty")
	}

	return u.UserStore.Update(ctx, user)
}

func (u *UserRouter) deleteUser(ctx context.Context, r *http.Request, deleteReq *UserDeleteRequest) error {
	if u.UserStore == nil {
		return router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		return err
	}

	return u.UserStore.Delete(ctx, id)
}

// UserDeleteRequest represents a request to delete a user (can be empty for path-based deletion)
type UserDeleteRequest struct {
	ID uuid.UUID `json:"id,omitempty"`
}
