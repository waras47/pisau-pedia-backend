package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(b, s)
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

type Review struct {
	ID            string       `db:"id"`
	// ProductID is nil for a "shop review" — feedback about the store
	// itself (shipping, service, packaging) rather than any one product.
	ProductID     *string      `db:"product_id"`
	CustomerName  string       `db:"customer_name"`
	CustomerEmail *string      `db:"customer_email"`
	Rating        uint         `db:"rating"`
	Content       string       `db:"content"`
	Photos        StringSlice  `db:"photos"`
	Status        ReviewStatus `db:"status"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`

	// Only populated by admin list queries (joined), not part of the
	// reviews table itself.
	ProductName *string `db:"product_name"`
	ProductSlug *string `db:"product_slug"`
}
