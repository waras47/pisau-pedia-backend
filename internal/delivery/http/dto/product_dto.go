package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type ProductSpecRequest struct {
	Label string `json:"label" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type CreateProductRequest struct {
	CategoryID     *string              `json:"category_id" validate:"omitempty,uuid"`
	Name           string               `json:"name" validate:"required,min=2"`
	Slug           string               `json:"slug" validate:"omitempty"`
	Description    *string              `json:"description"`
	Price          int64                `json:"price" validate:"required,min=0"`
	CompareAtPrice *int64               `json:"compare_at_price" validate:"omitempty,min=0"`
	Maker          *string              `json:"maker"`
	Badge          *string              `json:"badge" validate:"omitempty,oneof=new sale sold-out"`
	Stock          *uint                `json:"stock" validate:"omitempty,min=0"`
	Images         []string             `json:"images" validate:"omitempty,dive,url"`
	Specs          []ProductSpecRequest `json:"specs" validate:"omitempty,dive"`
	Highlights     []string             `json:"highlights" validate:"omitempty,dive,required"`
}

type UpdateProductRequest struct {
	CategoryID     *string `json:"category_id" validate:"omitempty,uuid"`
	Name           string  `json:"name" validate:"omitempty,min=2"`
	Description    *string `json:"description"`
	Price          int64   `json:"price" validate:"omitempty,min=0"`
	CompareAtPrice *int64  `json:"compare_at_price" validate:"omitempty,min=0"`
	Maker          *string `json:"maker"`
	Badge          *string `json:"badge" validate:"omitempty,oneof=new sale sold-out"`
	Stock          *uint   `json:"stock" validate:"omitempty,min=0"`
	IsActive       *bool   `json:"is_active"`
}

func (r CreateProductRequest) ToInput() usecase.ProductInput {
	input := usecase.ProductInput{
		CategoryID:     r.CategoryID,
		Name:           r.Name,
		Slug:           r.Slug,
		Description:    r.Description,
		Price:          r.Price,
		CompareAtPrice: r.CompareAtPrice,
		Maker:          r.Maker,
		Stock:          r.Stock,
		Images:         r.Images,
		Highlights:     r.Highlights,
	}
	if r.Badge != nil {
		badge := entity.Badge(*r.Badge)
		input.Badge = &badge
	}
	for _, s := range r.Specs {
		input.Specs = append(input.Specs, usecase.ProductSpecInput{Label: s.Label, Value: s.Value})
	}
	return input
}

func (r UpdateProductRequest) ToInput() usecase.ProductInput {
	input := usecase.ProductInput{
		CategoryID:     r.CategoryID,
		Name:           r.Name,
		Description:    r.Description,
		Price:          r.Price,
		CompareAtPrice: r.CompareAtPrice,
		Maker:          r.Maker,
		Stock:          r.Stock,
		IsActive:       r.IsActive,
	}
	if r.Badge != nil {
		badge := entity.Badge(*r.Badge)
		input.Badge = &badge
	}
	return input
}

type ProductListItemResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Price          int64   `json:"price"`
	CompareAtPrice *int64  `json:"compare_at_price,omitempty"`
	Currency       string  `json:"currency"`
	Maker          string  `json:"maker,omitempty"`
	Badge          string  `json:"badge,omitempty"`
	Category       string  `json:"category,omitempty"`
	Stock          uint    `json:"stock"`
	RatingAvg      float64 `json:"rating"`
	ReviewCount    uint    `json:"review_count"`
}

type ProductDetailResponse struct {
	ProductListItemResponse
	Description string   `json:"description,omitempty"`
	Images      []string `json:"images"`
	Specs       []ProductSpecRequest `json:"specs"`
	Highlights  []string `json:"highlights"`
}

func toProductListItem(p *entity.Product) ProductListItemResponse {
	resp := ProductListItemResponse{
		ID:             p.ID,
		Name:           p.Name,
		Slug:           p.Slug,
		Price:          p.Price,
		CompareAtPrice: p.CompareAtPrice,
		Currency:       p.Currency,
		Stock:          p.Stock,
		RatingAvg:      p.RatingAvg,
		ReviewCount:    p.ReviewCount,
	}
	if p.Maker != nil {
		resp.Maker = *p.Maker
	}
	if p.Badge != nil {
		resp.Badge = string(*p.Badge)
	}
	if p.CategoryName != nil {
		resp.Category = *p.CategoryName
	}
	return resp
}

func ToProductListItems(products []entity.Product) []ProductListItemResponse {
	out := make([]ProductListItemResponse, 0, len(products))
	for i := range products {
		out = append(out, toProductListItem(&products[i]))
	}
	return out
}

func ToProductDetailResponse(p *entity.Product) ProductDetailResponse {
	resp := ProductDetailResponse{ProductListItemResponse: toProductListItem(p)}
	if p.Description != nil {
		resp.Description = *p.Description
	}

	resp.Images = make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		resp.Images = append(resp.Images, img.URL)
	}

	resp.Specs = make([]ProductSpecRequest, 0, len(p.Specs))
	for _, s := range p.Specs {
		resp.Specs = append(resp.Specs, ProductSpecRequest{Label: s.Label, Value: s.Value})
	}

	resp.Highlights = make([]string, 0, len(p.Highlights))
	for _, h := range p.Highlights {
		resp.Highlights = append(resp.Highlights, h.Highlight)
	}

	return resp
}
