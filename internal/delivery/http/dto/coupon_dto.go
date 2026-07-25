package dto

import (
	"time"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CouponRequest struct {
	Code        string     `json:"code" validate:"required"`
	Type        string     `json:"type" validate:"required,oneof=percentage fixed free_shipping"`
	Value       int64      `json:"value" validate:"omitempty,min=0"`
	MinOrder    int64      `json:"min_order" validate:"omitempty,min=0"`
	MaxUses     *uint      `json:"max_uses" validate:"omitempty,min=1"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	IsActive    *bool      `json:"is_active"`
	ShowPopup   *bool      `json:"show_popup"`
	Description *string    `json:"description"`
}

func (r CouponRequest) ToInput() usecase.CouponInput {
	return usecase.CouponInput{
		Code:        r.Code,
		Type:        entity.CouponType(r.Type),
		Value:       r.Value,
		MinOrder:    r.MinOrder,
		MaxUses:     r.MaxUses,
		StartsAt:    r.StartsAt,
		EndsAt:      r.EndsAt,
		IsActive:    r.IsActive,
		ShowPopup:   r.ShowPopup,
		Description: r.Description,
	}
}

type ValidateCouponRequest struct {
	Code     string `json:"code" validate:"required"`
	Subtotal int64  `json:"subtotal" validate:"required,min=0"`
}

type CouponResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Type        string  `json:"type"`
	Value       int64   `json:"value"`
	MinOrder    int64   `json:"min_order"`
	MaxUses     *uint   `json:"max_uses,omitempty"`
	UsedCount   uint    `json:"used_count"`
	StartsAt    *string `json:"starts_at,omitempty"`
	EndsAt      *string `json:"ends_at,omitempty"`
	IsActive    bool    `json:"is_active"`
	ShowPopup   bool    `json:"show_popup"`
	Description string  `json:"description,omitempty"`
}

func ToCouponResponse(c *entity.Coupon) CouponResponse {
	resp := CouponResponse{
		ID:        c.ID,
		Code:      c.Code,
		Type:      string(c.Type),
		Value:     c.Value,
		MinOrder:  c.MinOrder,
		MaxUses:   c.MaxUses,
		UsedCount: c.UsedCount,
		IsActive:  c.IsActive,
		ShowPopup: c.ShowPopup,
	}
	if c.StartsAt != nil {
		s := c.StartsAt.Format("2006-01-02")
		resp.StartsAt = &s
	}
	if c.EndsAt != nil {
		s := c.EndsAt.Format("2006-01-02")
		resp.EndsAt = &s
	}
	if c.Description != nil {
		resp.Description = *c.Description
	}
	return resp
}

func ToCouponResponses(coupons []entity.Coupon) []CouponResponse {
	out := make([]CouponResponse, 0, len(coupons))
	for i := range coupons {
		out = append(out, ToCouponResponse(&coupons[i]))
	}
	return out
}

type PromoPopupResponse struct {
	Type        string `json:"type"`
	Value       int64  `json:"value"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

func ToPromoPopupResponse(c *entity.Coupon) PromoPopupResponse {
	resp := PromoPopupResponse{
		Type:  string(c.Type),
		Value: c.Value,
		Code:  c.Code,
	}
	if c.Description != nil {
		resp.Description = *c.Description
	}
	return resp
}

type ValidateCouponResponse struct {
	Valid          bool   `json:"valid"`
	DiscountAmount int64  `json:"discount_amount"`
	FreeShipping   bool   `json:"free_shipping"`
	Code           string `json:"code"`
}

func ToValidateCouponResponse(r *usecase.CouponValidationResult) ValidateCouponResponse {
	return ValidateCouponResponse{
		Valid:          true,
		DiscountAmount: r.DiscountAmount,
		FreeShipping:   r.FreeShipping,
		Code:           r.Coupon.Code,
	}
}
