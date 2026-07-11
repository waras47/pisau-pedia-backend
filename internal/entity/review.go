package entity

import "time"

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

type Review struct {
	ID            string       `db:"id"`
	ProductID     string       `db:"product_id"`
	CustomerName  string       `db:"customer_name"`
	CustomerEmail *string      `db:"customer_email"`
	Rating        uint         `db:"rating"`
	Content       string       `db:"content"`
	Status        ReviewStatus `db:"status"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`

	// Only populated by admin list queries (joined), not part of the
	// reviews table itself.
	ProductName *string `db:"product_name"`
	ProductSlug *string `db:"product_slug"`
}
