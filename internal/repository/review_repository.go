package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type ReviewFilter struct {
	Page        int
	PerPage     int
	ProductSlug string
	Status      string
	// Scope narrows to "product" (has a product_id) or "shop" (no
	// product_id) reviews. Empty means both.
	Scope string
}

type ProductRatingStats struct {
	RatingAvg   float64
	ReviewCount uint
}

type ReviewRepository interface {
	FindAll(ctx context.Context, filter ReviewFilter) ([]entity.Review, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Review, error)
	Create(ctx context.Context, review *entity.Review) error
	Update(ctx context.Context, review *entity.Review) error
	Delete(ctx context.Context, id string) error
	GetApprovedStatsByProduct(ctx context.Context, productID string) (*ProductRatingStats, error)
}
