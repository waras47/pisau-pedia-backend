package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type UserFilter struct {
	Page    int
	PerPage int
	Role    string
	Search  string
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	// FindByGoogleID looks up a user previously linked to this Google
	// account, returning ErrUserNotFound when none exists.
	FindByGoogleID(ctx context.Context, googleID string) (*entity.User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
}
