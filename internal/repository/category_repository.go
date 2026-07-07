package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
}
