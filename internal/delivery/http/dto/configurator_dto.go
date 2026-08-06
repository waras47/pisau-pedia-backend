package dto

import (
	"encoding/json"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

// --- Shape ---

type CreateShapeRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Category    string  `json:"category" validate:"required"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateShapeRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2"`
	Category    string  `json:"category"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
	SortOrder   int     `json:"sort_order"`
}

func (r CreateShapeRequest) ToInput() usecase.ConfiguratorShapeInput {
	return usecase.ConfiguratorShapeInput{Name: r.Name, Category: r.Category, Description: r.Description, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

func (r UpdateShapeRequest) ToInput() usecase.ConfiguratorShapeInput {
	return usecase.ConfiguratorShapeInput{Name: r.Name, Category: r.Category, Description: r.Description, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

type ShapeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

func ToShapeResponse(s *entity.ConfiguratorShape) ShapeResponse {
	r := ShapeResponse{ID: s.ID, Name: s.Name, Category: s.Category, SortOrder: s.SortOrder}
	if s.Description != nil {
		r.Description = *s.Description
	}
	if s.ImageURL != nil {
		r.ImageURL = *s.ImageURL
	}
	return r
}

func ToShapeResponses(shapes []entity.ConfiguratorShape) []ShapeResponse {
	out := make([]ShapeResponse, 0, len(shapes))
	for i := range shapes {
		out = append(out, ToShapeResponse(&shapes[i]))
	}
	return out
}

// --- Blade ---

type CreateBladeRequest struct {
	ShapeID        string            `json:"shape_id" validate:"required"`
	Name           string            `json:"name" validate:"required,min=2"`
	Steel          string            `json:"steel" validate:"required"`
	LengthMm       int               `json:"length_mm" validate:"required,min=1"`
	Price          float64           `json:"price" validate:"min=0"`
	CompareAtPrice *float64          `json:"compare_at_price"`
	Description    *string           `json:"description"`
	Specifications map[string]string `json:"specifications"`
	ImageURL       *string           `json:"image_url" validate:"omitempty,url"`
	SortOrder      int               `json:"sort_order"`
}

type UpdateBladeRequest struct {
	ShapeID        string            `json:"shape_id"`
	Name           string            `json:"name" validate:"omitempty,min=2"`
	Steel          string            `json:"steel"`
	LengthMm       int               `json:"length_mm"`
	Price          float64           `json:"price"`
	CompareAtPrice *float64          `json:"compare_at_price"`
	Description    *string           `json:"description"`
	Specifications map[string]string `json:"specifications"`
	ImageURL       *string           `json:"image_url" validate:"omitempty,url"`
	SortOrder      int               `json:"sort_order"`
}

func (r CreateBladeRequest) ToInput() usecase.ConfiguratorBladeInput {
	return usecase.ConfiguratorBladeInput{
		ShapeID: r.ShapeID, Name: r.Name, Steel: r.Steel, LengthMm: r.LengthMm,
		Price: r.Price, CompareAtPrice: r.CompareAtPrice, Description: r.Description,
		Specifications: r.Specifications, ImageURL: r.ImageURL, SortOrder: r.SortOrder,
	}
}

func (r UpdateBladeRequest) ToInput() usecase.ConfiguratorBladeInput {
	return usecase.ConfiguratorBladeInput{
		ShapeID: r.ShapeID, Name: r.Name, Steel: r.Steel, LengthMm: r.LengthMm,
		Price: r.Price, CompareAtPrice: r.CompareAtPrice, Description: r.Description,
		Specifications: r.Specifications, ImageURL: r.ImageURL, SortOrder: r.SortOrder,
	}
}

type BladeResponse struct {
	ID             string            `json:"id"`
	ShapeID        string            `json:"shape_id"`
	Name           string            `json:"name"`
	Steel          string            `json:"steel"`
	LengthMm       int               `json:"length_mm"`
	Price          float64           `json:"price"`
	CompareAtPrice *float64          `json:"compare_at_price,omitempty"`
	Description    string            `json:"description,omitempty"`
	Specifications map[string]string `json:"specifications,omitempty"`
	ImageURL       string            `json:"image_url,omitempty"`
	SortOrder      int               `json:"sort_order"`
}

func ToBladeResponse(b *entity.ConfiguratorBlade) BladeResponse {
	r := BladeResponse{
		ID: b.ID, ShapeID: b.ShapeID, Name: b.Name, Steel: b.Steel,
		LengthMm: b.LengthMm, Price: b.Price, CompareAtPrice: b.CompareAtPrice, SortOrder: b.SortOrder,
	}
	if b.Description != nil {
		r.Description = *b.Description
	}
	if b.ImageURL != nil {
		r.ImageURL = *b.ImageURL
	}
	if len(b.Specifications) > 0 {
		_ = json.Unmarshal(b.Specifications, &r.Specifications)
	}
	return r
}

func ToBladeResponses(blades []entity.ConfiguratorBlade) []BladeResponse {
	out := make([]BladeResponse, 0, len(blades))
	for i := range blades {
		out = append(out, ToBladeResponse(&blades[i]))
	}
	return out
}

// --- Handle ---

type CreateHandleRequest struct {
	Name       string  `json:"name" validate:"required,min=2"`
	Material   string  `json:"material" validate:"required"`
	PriceDelta float64 `json:"price_delta"`
	ImageURL   *string `json:"image_url" validate:"omitempty,url"`
	SortOrder  int     `json:"sort_order"`
}

type UpdateHandleRequest struct {
	Name       string  `json:"name" validate:"omitempty,min=2"`
	Material   string  `json:"material"`
	PriceDelta float64 `json:"price_delta"`
	ImageURL   *string `json:"image_url" validate:"omitempty,url"`
	SortOrder  int     `json:"sort_order"`
}

func (r CreateHandleRequest) ToInput() usecase.ConfiguratorHandleInput {
	return usecase.ConfiguratorHandleInput{Name: r.Name, Material: r.Material, PriceDelta: r.PriceDelta, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

func (r UpdateHandleRequest) ToInput() usecase.ConfiguratorHandleInput {
	return usecase.ConfiguratorHandleInput{Name: r.Name, Material: r.Material, PriceDelta: r.PriceDelta, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

type HandleResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Material   string  `json:"material"`
	PriceDelta float64 `json:"price_delta"`
	ImageURL   string  `json:"image_url,omitempty"`
	SortOrder  int     `json:"sort_order"`
}

func ToHandleResponse(h *entity.ConfiguratorHandle) HandleResponse {
	r := HandleResponse{ID: h.ID, Name: h.Name, Material: h.Material, PriceDelta: h.PriceDelta, SortOrder: h.SortOrder}
	if h.ImageURL != nil {
		r.ImageURL = *h.ImageURL
	}
	return r
}

func ToHandleResponses(handles []entity.ConfiguratorHandle) []HandleResponse {
	out := make([]HandleResponse, 0, len(handles))
	for i := range handles {
		out = append(out, ToHandleResponse(&handles[i]))
	}
	return out
}

// --- Accessory ---

type CreateAccessoryRequest struct {
	Name     string  `json:"name" validate:"required,min=2"`
	Price    float64 `json:"price" validate:"min=0"`
	ImageURL *string `json:"image_url" validate:"omitempty,url"`
	SortOrder int    `json:"sort_order"`
}

type UpdateAccessoryRequest struct {
	Name     string  `json:"name" validate:"omitempty,min=2"`
	Price    float64 `json:"price"`
	ImageURL *string `json:"image_url" validate:"omitempty,url"`
	SortOrder int    `json:"sort_order"`
}

func (r CreateAccessoryRequest) ToInput() usecase.ConfiguratorAccessoryInput {
	return usecase.ConfiguratorAccessoryInput{Name: r.Name, Price: r.Price, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

func (r UpdateAccessoryRequest) ToInput() usecase.ConfiguratorAccessoryInput {
	return usecase.ConfiguratorAccessoryInput{Name: r.Name, Price: r.Price, ImageURL: r.ImageURL, SortOrder: r.SortOrder}
}

type AccessoryResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	ImageURL  string  `json:"image_url,omitempty"`
	SortOrder int     `json:"sort_order"`
}

func ToAccessoryResponse(a *entity.ConfiguratorAccessory) AccessoryResponse {
	r := AccessoryResponse{ID: a.ID, Name: a.Name, Price: a.Price, SortOrder: a.SortOrder}
	if a.ImageURL != nil {
		r.ImageURL = *a.ImageURL
	}
	return r
}

func ToAccessoryResponses(accessories []entity.ConfiguratorAccessory) []AccessoryResponse {
	out := make([]AccessoryResponse, 0, len(accessories))
	for i := range accessories {
		out = append(out, ToAccessoryResponse(&accessories[i]))
	}
	return out
}
