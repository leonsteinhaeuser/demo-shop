package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonsteinhaeuser/demo-shop/internal/handlers"
	"github.com/leonsteinhaeuser/demo-shop/internal/router"
	"github.com/prometheus/client_golang/prometheus"
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

	IsAdmin bool `json:"is_admin"`
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
	UserStore               UserStore
	processedCreateRequests prometheus.Counter
	processedCreateFailures prometheus.Counter
	processedListRequests   prometheus.Counter
	processedListFailures   prometheus.Counter
	processedGetRequests    prometheus.Counter
	processedGetFailures    prometheus.Counter
	processedUpdateRequests prometheus.Counter
	processedUpdateFailures prometheus.Counter
	processedDeleteRequests prometheus.Counter
	processedDeleteFailures prometheus.Counter
}

func NewUserRouter(userStore UserStore) *UserRouter {
	return &UserRouter{
		UserStore: userStore,
		processedCreateRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_create_processed_requests_total",
				Help: "Total number of user create requests",
			},
		),
		processedCreateFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_create_processed_failures_total",
				Help: "Total number of user create request failures",
			},
		),
		processedListRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_list_processed_requests_total",
				Help: "Total number of user list requests",
			},
		),
		processedListFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_list_processed_failures_total",
				Help: "Total number of user list request failures",
			},
		),
		processedGetRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_get_processed_requests_total",
				Help: "Total number of user get requests",
			},
		),
		processedGetFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_get_processed_failures_total",
				Help: "Total number of user get request failures",
			},
		),
		processedUpdateRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_update_processed_requests_total",
				Help: "Total number of user update requests",
			},
		),
		processedUpdateFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_update_processed_failures_total",
				Help: "Total number of user update request failures",
			},
		),
		processedDeleteRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_delete_processed_requests_total",
				Help: "Total number of user delete requests",
			},
		),
		processedDeleteFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_delete_processed_failures_total",
				Help: "Total number of user delete request failures",
			},
		),
	}
}

func (u *UserRouter) GetApiVersion() string {
	return version
}

func (u *UserRouter) GetGroup() string {
	return group
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
	u.processedCreateRequests.Inc()

	if u.UserStore == nil {
		u.processedCreateFailures.Inc()
		return router.ErrObjectStorageNotImplemented
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if user.Password == nil || *user.Password == "" {
		u.processedCreateFailures.Inc()
		return errors.New("password is required")
	}
	// ensure password meets security requirements (e.g., length, complexity) here if needed
	// For simplicity, let's say it must be at least 12 characters long
	if len(*user.Password) < 12 {
		u.processedCreateFailures.Inc()
		return errors.New("password must be at least 12 characters long")
	}
	if user.Username == nil || *user.Username == "" {
		u.processedCreateFailures.Inc()
		return errors.New("username cannot be empty")
	}
	if user.Email == nil || *user.Email == "" {
		u.processedCreateFailures.Inc()
		return errors.New("email cannot be empty")
	}

	err := u.UserStore.Create(ctx, user)
	if err != nil {
		u.processedCreateFailures.Inc()
		return err
	}
	return nil
}

func (u *UserRouter) listUsers(ctx context.Context, r *http.Request, filters handlers.FilterObjectList) ([]User, error) {
	u.processedListRequests.Inc()

	if u.UserStore == nil {
		u.processedListFailures.Inc()
		return nil, router.ErrObjectStorageNotImplemented
	}

	users, err := u.UserStore.List(ctx, filters.Page, filters.Limit)
	if err != nil {
		u.processedListFailures.Inc()
		return nil, err
	}
	return users, nil
}

func (u *UserRouter) getUser(ctx context.Context, r *http.Request) (*User, error) {
	u.processedGetRequests.Inc()

	if u.UserStore == nil {
		u.processedGetFailures.Inc()
		return nil, router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		u.processedGetFailures.Inc()
		return nil, err
	}

	user, err := u.UserStore.Get(ctx, id)
	if err != nil {
		u.processedGetFailures.Inc()
		return nil, err
	}
	return user, nil
}

func (u *UserRouter) updateUser(ctx context.Context, r *http.Request, user *UserModificationRequest) error {
	u.processedUpdateRequests.Inc()

	if u.UserStore == nil {
		u.processedUpdateFailures.Inc()
		return router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		u.processedUpdateFailures.Inc()
		return err
	}
	if user.ID != id {
		u.processedUpdateFailures.Inc()
		return errors.New("ID in path does not match user ID")
	}

	user.UpdatedAt = time.Now()

	// if the password is not provided, we do not update it
	if user.Password != nil {
		if *user.Password == "" {
			u.processedUpdateFailures.Inc()
			return errors.New("password is required")
		}
		if len(*user.Password) < 12 {
			u.processedUpdateFailures.Inc()
			return errors.New("password must be at least 12 characters long")
		}

	}

	if user.Username != nil && *user.Username == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("username cannot be empty")
	}
	if user.Email != nil && *user.Email == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("email cannot be empty")
	}
	if user.PreferredName != nil && *user.PreferredName == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("preferred_name cannot be empty")
	}
	if user.GivenName != nil && *user.GivenName == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("given_name cannot be empty")
	}
	if user.FamilyName != nil && *user.FamilyName == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("family_name cannot be empty")
	}
	if user.Locale != nil && *user.Locale == "" {
		u.processedUpdateFailures.Inc()
		return errors.New("locale cannot be empty")
	}

	err = u.UserStore.Update(ctx, user)
	if err != nil {
		u.processedUpdateFailures.Inc()
		return err
	}
	return nil
}

func (u *UserRouter) deleteUser(ctx context.Context, r *http.Request, deleteReq *UserDeleteRequest) error {
	u.processedDeleteRequests.Inc()

	if u.UserStore == nil {
		u.processedDeleteFailures.Inc()
		return router.ErrObjectStorageNotImplemented
	}

	id, err := handlers.GetUUIDFromPathValue(r, "id")
	if err != nil {
		u.processedDeleteFailures.Inc()
		return err
	}

	err = u.UserStore.Delete(ctx, id)
	if err != nil {
		u.processedDeleteFailures.Inc()
		return err
	}
	return nil
}

// UserDeleteRequest represents a request to delete a user (can be empty for path-based deletion)
type UserDeleteRequest struct {
	ID uuid.UUID `json:"id,omitempty"`
}
