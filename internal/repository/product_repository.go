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

type CategoryStockPoint struct {
	CategoryName string `db:"category_name"`
	ProductCount int64  `db:"product_count"`
	TotalStock   int64  `db:"total_stock"`
}

type InventorySummary struct {
	TotalProducts     int64
	TotalStockValue   int64
	LowStockCount     int64
	OutOfStockCount   int64
	CategoryBreakdown []CategoryStockPoint
}

type ProductRepository interface {
	FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	GetInventorySummary(ctx context.Context, lowStockThreshold uint) (*InventorySummary, error)
	UpdateRatingStats(ctx context.Context, productID string, ratingAvg float64, reviewCount uint) error
}
