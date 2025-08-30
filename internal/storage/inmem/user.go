package inmem

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	apiv1 "github.com/leonsteinhaeuser/demo-shop/api/v1"
	"github.com/leonsteinhaeuser/demo-shop/internal/utils"
)

var (
	defaultUser    = uuid.New()
	defaultRegUser = uuid.New()

	_ apiv1.UserStore = (*UserInMemStorage)(nil)
)

type UserInMemStorage struct {
	users map[string]*apiv1.UserModificationRequest
}

func NewUserInMemStorage() *UserInMemStorage {
	return &UserInMemStorage{
		users: map[string]*apiv1.UserModificationRequest{
			defaultUser.String(): {
				User: apiv1.User{
					ID:        defaultUser,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),

					Username:      utils.StringPtr("root"),
					Email:         utils.StringPtr("root@localhost"),
					EmailVerified: true,
					PreferredName: utils.StringPtr("Root User"),
					GivenName:     utils.StringPtr("Root"),
					FamilyName:    utils.StringPtr("User"),
					Locale:        utils.StringPtr("en/US"),

					IsAdmin: true,
				},
				Password: utils.StringPtr("root"),
			},
			defaultRegUser.String(): {
				User: apiv1.User{
					ID:        defaultRegUser,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),

					Username:      utils.StringPtr("user"),
					Email:         utils.StringPtr("user@localhost"),
					EmailVerified: true,
					PreferredName: utils.StringPtr("Regular User"),
					GivenName:     utils.StringPtr("Regular"),
					FamilyName:    utils.StringPtr("User"),
					Locale:        utils.StringPtr("en/US"),

					IsAdmin: false,
				},
				Password: utils.StringPtr("userpassword"),
			},
		},
	}
}

func (s *UserInMemStorage) Create(ctx context.Context, user *apiv1.UserModificationRequest) error {
	for {
		id := uuid.New()
		if _, exists := s.users[id.String()]; exists {
			continue
		}
		user.ID = id
		break
	}
	s.users[user.ID.String()] = user
	return nil
}

func (s *UserInMemStorage) List(ctx context.Context, page, limit int) ([]apiv1.User, error) {
	var users []apiv1.User
	for _, user := range s.users {
		users = append(users, user.User)
	}
	return users, nil
}

func (s *UserInMemStorage) Get(ctx context.Context, id uuid.UUID) (*apiv1.User, error) {
	user, exists := s.users[id.String()]
	if !exists {
		return nil, errors.New("user not found")
	}
	return &user.User, nil
}

func (s *UserInMemStorage) Update(ctx context.Context, user *apiv1.UserModificationRequest) error {
	existingUser, exists := s.users[user.ID.String()]
	if !exists {
		return errors.New("user not found")
	}
	existingUser.User = user.User
	return nil
}

func (s *UserInMemStorage) Delete(ctx context.Context, id uuid.UUID) error {
	delete(s.users, id.String())
	return nil
}
