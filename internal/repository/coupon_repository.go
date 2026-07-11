package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type CouponFilter struct {
	Page    int
	PerPage int
}

type CouponRepository interface {
	FindAll(ctx context.Context, filter CouponFilter) ([]entity.Coupon, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Coupon, error)
	FindByCode(ctx context.Context, code string) (*entity.Coupon, error)
	Create(ctx context.Context, coupon *entity.Coupon) error
	Update(ctx context.Context, coupon *entity.Coupon) error
	Delete(ctx context.Context, id string) error
	IncrementUsage(ctx context.Context, id string) error
}
