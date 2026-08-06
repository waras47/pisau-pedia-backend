package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type CollectionRepository interface {
	FindAll(ctx context.Context) ([]entity.Collection, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Collection, error)
	FindByID(ctx context.Context, id string) (*entity.Collection, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, collection *entity.Collection) error
	Update(ctx context.Context, collection *entity.Collection) error
	Delete(ctx context.Context, id string) error
	SetCategories(ctx context.Context, collectionID string, categoryIDs []string) error
}
