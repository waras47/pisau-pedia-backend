package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type AddressRepository interface {
	Create(ctx context.Context, address *entity.Address) error
	FindByUserID(ctx context.Context, userID string) ([]entity.Address, error)
	FindByID(ctx context.Context, id string) (*entity.Address, error)
	Update(ctx context.Context, address *entity.Address) error
	Delete(ctx context.Context, id string) error
	UnsetDefaultForUser(ctx context.Context, userID string) error
}
