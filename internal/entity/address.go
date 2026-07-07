package entity

import "time"

type Address struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Label       *string   `db:"label"`
	FullName    string    `db:"full_name"`
	Phone       *string   `db:"phone"`
	AddressLine string    `db:"address_line"`
	City        string    `db:"city"`
	Province    *string   `db:"province"`
	PostalCode  string    `db:"postal_code"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
}
