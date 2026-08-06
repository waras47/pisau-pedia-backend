package entity

import "time"

type Collection struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Description *string   `db:"description"`
	ImageURL    *string   `db:"image_url"`
	SortOrder   int       `db:"sort_order"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	Categories []Category `db:"-"`
}
