package entity

import "time"

type SiteContent struct {
	ID        string    `db:"id"`
	Key       string    `db:"key"`
	Value     string    `db:"value"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
