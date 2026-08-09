package entity

import "time"

type SitePromo struct {
	ID               string    `db:"id"`
	Title            string    `db:"title"`
	Description      *string   `db:"description"`
	DiscountPercent  int       `db:"discount_percent"`
	PopupImage       *string   `db:"popup_image"`
	ApplyToAll       bool      `db:"apply_to_all"`
	StartDate        time.Time `db:"start_date"`
	EndDate          time.Time `db:"end_date"`
	IsActive         bool      `db:"is_active"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	ProductIDs       []string  `db:"-"`
}
