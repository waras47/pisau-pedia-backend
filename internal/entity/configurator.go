package entity

import (
	"encoding/json"
	"time"
)

type ConfiguratorShape struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Category    string    `db:"category"`
	Description *string   `db:"description"`
	ImageURL    *string   `db:"image_url"`
	SortOrder   int       `db:"sort_order"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ConfiguratorBlade struct {
	ID             string          `db:"id"`
	ShapeID        string          `db:"shape_id"`
	Name           string          `db:"name"`
	Steel          string          `db:"steel"`
	LengthMm       int             `db:"length_mm"`
	Price          float64         `db:"price"`
	CompareAtPrice *float64        `db:"compare_at_price"`
	Description    *string         `db:"description"`
	Specifications json.RawMessage `db:"specifications"`
	ImageURL       *string         `db:"image_url"`
	SortOrder      int             `db:"sort_order"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
}

type ConfiguratorHandle struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Material   string    `db:"material"`
	PriceDelta float64   `db:"price_delta"`
	ImageURL   *string   `db:"image_url"`
	SortOrder  int       `db:"sort_order"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type ConfiguratorAccessory struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Price     float64   `db:"price"`
	ImageURL  *string   `db:"image_url"`
	SortOrder int       `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
