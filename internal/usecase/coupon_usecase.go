package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

var (
	ErrCouponInvalid        = errors.New("coupon is invalid, expired, or fully redeemed")
	ErrCouponMinOrderNotMet = errors.New("order does not meet the coupon's minimum amount")
)

type CouponInput struct {
	Code        string
	Type        entity.CouponType
	Value       int64
	MinOrder    int64
	MaxUses     *uint
	StartsAt    *time.Time
	EndsAt      *time.Time
	IsActive    *bool
	Description *string
}

type CouponListResult struct {
	Coupons    []entity.Coupon
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

type CouponValidationResult struct {
	Coupon         *entity.Coupon
	DiscountAmount int64
	FreeShipping   bool
}

type CouponUsecase struct {
	couponRepo repository.CouponRepository
}

func NewCouponUsecase(couponRepo repository.CouponRepository) *CouponUsecase {
	return &CouponUsecase{couponRepo: couponRepo}
}

func (u *CouponUsecase) ListCoupons(ctx context.Context, page, perPage int) (*CouponListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	coupons, total, err := u.couponRepo.FindAll(ctx, repository.CouponFilter{Page: page, PerPage: perPage})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &CouponListResult{Coupons: coupons, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *CouponUsecase) CreateCoupon(ctx context.Context, input CouponInput) (*entity.Coupon, error) {
	coupon := &entity.Coupon{
		ID:          uuid.New().String(),
		Code:        input.Code,
		Type:        input.Type,
		Value:       input.Value,
		MinOrder:    input.MinOrder,
		MaxUses:     input.MaxUses,
		StartsAt:    input.StartsAt,
		EndsAt:      input.EndsAt,
		IsActive:    true,
		Description: input.Description,
	}
	if input.IsActive != nil {
		coupon.IsActive = *input.IsActive
	}
	if err := u.couponRepo.Create(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (u *CouponUsecase) UpdateCoupon(ctx context.Context, id string, input CouponInput) (*entity.Coupon, error) {
	coupon, err := u.couponRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	coupon.Code = input.Code
	coupon.Type = input.Type
	coupon.Value = input.Value
	coupon.MinOrder = input.MinOrder
	coupon.MaxUses = input.MaxUses
	coupon.StartsAt = input.StartsAt
	coupon.EndsAt = input.EndsAt
	coupon.Description = input.Description
	if input.IsActive != nil {
		coupon.IsActive = *input.IsActive
	}

	if err := u.couponRepo.Update(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (u *CouponUsecase) DeleteCoupon(ctx context.Context, id string) error {
	if _, err := u.couponRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.couponRepo.Delete(ctx, id)
}

// ValidateCoupon is shared by the storefront "apply coupon" preview endpoint
// and OrderUsecase.CreateOrder — both must agree on what makes a coupon
// usable, so the rule lives in exactly one place.
func (u *CouponUsecase) ValidateCoupon(ctx context.Context, code string, subtotal int64) (*CouponValidationResult, error) {
	coupon, err := u.couponRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if !coupon.IsActive {
		return nil, ErrCouponInvalid
	}
	if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
		return nil, ErrCouponInvalid
	}
	if coupon.EndsAt != nil && now.After(*coupon.EndsAt) {
		return nil, ErrCouponInvalid
	}
	if coupon.MaxUses != nil && coupon.UsedCount >= *coupon.MaxUses {
		return nil, ErrCouponInvalid
	}
	if subtotal < coupon.MinOrder {
		return nil, ErrCouponMinOrderNotMet
	}

	result := &CouponValidationResult{Coupon: coupon}
	switch coupon.Type {
	case entity.CouponTypePercentage:
		result.DiscountAmount = subtotal * coupon.Value / 100
	case entity.CouponTypeFixed:
		if coupon.Value > subtotal {
			result.DiscountAmount = subtotal
		} else {
			result.DiscountAmount = coupon.Value
		}
	case entity.CouponTypeFreeShipping:
		result.FreeShipping = true
	}

	return result, nil
}
