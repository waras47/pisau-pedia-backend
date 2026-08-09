package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type SitePromoRepository interface {
	FindAll(ctx context.Context) ([]entity.SitePromo, error)
	FindByID(ctx context.Context, id string) (*entity.SitePromo, error)
	FindActive(ctx context.Context) (*entity.SitePromo, error)
	Create(ctx context.Context, promo *entity.SitePromo) error
	Update(ctx context.Context, promo *entity.SitePromo) error
	Delete(ctx context.Context, id string) error
	SetProductIDs(ctx context.Context, promoID string, productIDs []string) error
	GetProductIDs(ctx context.Context, promoID string) ([]string, error)
}
