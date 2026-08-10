package entity

import "time"

type Badge string

const (
	BadgeNew     Badge = "new"
	BadgeSoldOut Badge = "sold-out"
)

type Product struct {
	ID               string    `db:"id"`
	CategoryID       *string   `db:"category_id"`
	CategoryName     *string   `db:"category_name"`
	Name             string    `db:"name"`
	Slug             string    `db:"slug"`
	Description      *string   `db:"description"`
	DescriptionEN    *string   `db:"description_en"`
	CareInstructions *string   `db:"care_instructions"`
	Price            int64     `db:"price"`
	CompareAtPrice   *int64    `db:"compare_at_price"`
	Currency         string    `db:"currency"`
	Maker            *string   `db:"maker"`
	Badge            *Badge    `db:"badge"`
	Stock            uint      `db:"stock"`
	Weight           uint      `db:"weight"`
	RatingAvg        float64   `db:"rating_avg"`
	ReviewCount      uint      `db:"review_count"`
	IsActive         bool      `db:"is_active"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`

	// Only populated by GetBySlug — listing queries skip these on purpose
	// to avoid N+1 joins against three child tables for a page of results.
	Images     []ProductImage     `db:"-"`
	Specs      []ProductSpec      `db:"-"`
	Highlights []ProductHighlight `db:"-"`

	// Image is the first product image, populated separately by listing
	// queries (single batched query, not a join) so the list view has a
	// thumbnail without pulling the full Images/Specs/Highlights payload.
	Image *string `db:"-"`
}

type ProductImage struct {
	ID        string  `db:"id"`
	ProductID string  `db:"product_id"`
	URL       string  `db:"url"`
	AltText   *string `db:"alt_text"`
	// Angle tags one of the 4 admin-set product photo slots (front/back/
	// side/top) — nil for images added to the general gallery instead.
	Angle     *string `db:"angle"`
	SortOrder uint    `db:"sort_order"`
}

type ProductSpec struct {
	ID        string `db:"id"`
	ProductID string `db:"product_id"`
	Label     string `db:"label"`
	Value     string `db:"value"`
	SortOrder uint   `db:"sort_order"`
}

type ProductHighlight struct {
	ID        string `db:"id"`
	ProductID string `db:"product_id"`
	Highlight string `db:"highlight"`
	SortOrder uint   `db:"sort_order"`
}
