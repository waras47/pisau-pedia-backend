package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type ProductFilter struct {
	Page         int
	PerPage      int
	CategorySlug string
	Search       string
	Sort         string
}

type ProductRepository interface {
	FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
}
