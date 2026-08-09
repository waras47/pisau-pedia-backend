package entity

import "time"

type PostCategory struct {
	ID        string    `db:"id"`
	Slug      string    `db:"slug"`
	NameID    string    `db:"name_id"`
	NameEN    string    `db:"name_en"`
	DescID    *string   `db:"desc_id"`
	DescEN    *string   `db:"desc_en"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Post struct {
	ID             string     `db:"id"`
	Slug           string     `db:"slug"`
	CategoryID     string     `db:"category_id"`
	TitleID        string     `db:"title_id"`
	TitleEN        string     `db:"title_en"`
	ExcerptID      *string    `db:"excerpt_id"`
	ExcerptEN      *string    `db:"excerpt_en"`
	ContentID      *string    `db:"content_id"`
	ContentEN      *string    `db:"content_en"`
	Image          *string    `db:"image"`
	ReadingMinutes int        `db:"reading_minutes"`
	Status         string     `db:"status"`
	PublishedAt    *time.Time `db:"published_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`

	// Joined fields
	CategorySlug string `db:"category_slug"`
	CategoryName string `db:"category_name"`
}
