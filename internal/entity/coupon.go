package entity

import "time"

type CouponType string

const (
	CouponTypePercentage   CouponType = "percentage"
	CouponTypeFixed        CouponType = "fixed"
	CouponTypeFreeShipping CouponType = "free_shipping"
)

type Coupon struct {
	ID          string     `db:"id"`
	Code        string     `db:"code"`
	Type        CouponType `db:"type"`
	Value       int64      `db:"value"`
	MinOrder    int64      `db:"min_order"`
	MaxUses     *uint      `db:"max_uses"`
	UsedCount   uint       `db:"used_count"`
	StartsAt    *time.Time `db:"starts_at"`
	EndsAt      *time.Time `db:"ends_at"`
	IsActive    bool       `db:"is_active"`
	ShowPopup   bool       `db:"show_popup"`
	Description *string    `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
