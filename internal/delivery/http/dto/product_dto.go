package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type ProductSpecRequest struct {
	Label string `json:"label" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// ProductImageRequest lets an image optionally be tagged with one of the 4
// admin photo-capture angles — untagged images just join the general
// gallery.
type ProductImageRequest struct {
	URL   string  `json:"url" validate:"required,url"`
	Angle *string `json:"angle" validate:"omitempty,oneof=front back side top"`
}

type CreateProductRequest struct {
	CategoryID       *string               `json:"category_id" validate:"omitempty,uuid"`
	Name             string                `json:"name" validate:"required,min=2"`
	Slug             string                `json:"slug" validate:"omitempty"`
	Description      *string               `json:"description"`
	DescriptionEN    *string               `json:"description_en"`
	CareInstructions *string               `json:"care_instructions"`
	Price            int64                 `json:"price" validate:"required,min=0"`
	CompareAtPrice   *int64                `json:"compare_at_price" validate:"omitempty,min=0"`
	Maker            *string               `json:"maker"`
	Badge            *string               `json:"badge" validate:"omitempty,oneof=new sold-out"`
	Stock            *uint                 `json:"stock" validate:"omitempty,min=0"`
	Weight           *uint                 `json:"weight" validate:"omitempty,min=0"`
	Images           []ProductImageRequest `json:"images" validate:"omitempty,dive"`
	Specs            []ProductSpecRequest  `json:"specs" validate:"omitempty,dive"`
	Highlights       []string              `json:"highlights" validate:"omitempty,dive,required"`
}

type UpdateProductRequest struct {
	CategoryID       *string               `json:"category_id" validate:"omitempty,uuid"`
	Name             string                `json:"name" validate:"omitempty,min=2"`
	Description      *string               `json:"description"`
	DescriptionEN    *string               `json:"description_en"`
	CareInstructions *string               `json:"care_instructions"`
	Price            int64                 `json:"price" validate:"omitempty,min=0"`
	CompareAtPrice   *int64                `json:"compare_at_price" validate:"omitempty,min=0"`
	Maker            *string               `json:"maker"`
	Badge            *string               `json:"badge" validate:"omitempty,oneof=new sold-out"`
	Stock            *uint                 `json:"stock" validate:"omitempty,min=0"`
	Weight           *uint                 `json:"weight" validate:"omitempty,min=0"`
	IsActive         *bool                 `json:"is_active"`
	Images           []ProductImageRequest `json:"images" validate:"omitempty,dive"`
	Specs            []ProductSpecRequest  `json:"specs" validate:"omitempty,dive"`
	Highlights       []string              `json:"highlights" validate:"omitempty,dive,required"`
}

// toProductImageInputs preserves nil vs. empty-slice: UpdateProduct treats
// a nil Images as "field omitted, leave as-is" (see usecase.ProductInput),
// so an omitted `images` key in the request must stay nil here too, not
// become an empty slice that would wipe out existing images.
func toProductImageInputs(images []ProductImageRequest) []usecase.ProductImageInput {
	if images == nil {
		return nil
	}
	out := make([]usecase.ProductImageInput, 0, len(images))
	for _, img := range images {
		out = append(out, usecase.ProductImageInput{URL: img.URL, Angle: img.Angle})
	}
	return out
}

func (r CreateProductRequest) ToInput() usecase.ProductInput {
	input := usecase.ProductInput{
		CategoryID:       r.CategoryID,
		Name:             r.Name,
		Slug:             r.Slug,
		Description:      r.Description,
		DescriptionEN:    r.DescriptionEN,
		CareInstructions: r.CareInstructions,
		Price:            r.Price,
		CompareAtPrice:   r.CompareAtPrice,
		Maker:            r.Maker,
		Stock:            r.Stock,
		Weight:           r.Weight,
		Images:           toProductImageInputs(r.Images),
		Highlights:       r.Highlights,
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
		CategoryID:       r.CategoryID,
		Name:             r.Name,
		Description:      r.Description,
		DescriptionEN:    r.DescriptionEN,
		CareInstructions: r.CareInstructions,
		Price:            r.Price,
		CompareAtPrice:   r.CompareAtPrice,
		Maker:            r.Maker,
		Stock:            r.Stock,
		Weight:           r.Weight,
		IsActive:         r.IsActive,
		Images:           toProductImageInputs(r.Images),
		Highlights:       r.Highlights,
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
	Weight         uint    `json:"weight"`
	RatingAvg      float64 `json:"rating"`
	ReviewCount    uint    `json:"review_count"`
	Image          string  `json:"image,omitempty"`
}

// ProductAngleImagesResponse surfaces the 4 admin photo-capture slots
// individually so the storefront can render a fixed front/back/side/top
// switcher instead of guessing from the general gallery order.
type ProductAngleImagesResponse struct {
	Front string `json:"front,omitempty"`
	Back  string `json:"back,omitempty"`
	Side  string `json:"side,omitempty"`
	Top   string `json:"top,omitempty"`
}

type ProductDetailResponse struct {
	ProductListItemResponse
	Description      string                     `json:"description,omitempty"`
	DescriptionEN    string                     `json:"description_en,omitempty"`
	CareInstructions string                     `json:"care_instructions,omitempty"`
	Images           []string                   `json:"images"`
	AngleImages      ProductAngleImagesResponse `json:"angle_images"`
	Specs            []ProductSpecRequest       `json:"specs"`
	Highlights       []string                   `json:"highlights"`
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
		Weight:         p.Weight,
		RatingAvg:      p.RatingAvg,
		ReviewCount:    p.ReviewCount,
	}
	if p.Image != nil {
		resp.Image = *p.Image
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
	if p.DescriptionEN != nil {
		resp.DescriptionEN = *p.DescriptionEN
	}
	if p.CareInstructions != nil {
		resp.CareInstructions = *p.CareInstructions
	}

	resp.Images = make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		resp.Images = append(resp.Images, img.URL)
		if img.Angle == nil {
			continue
		}
		switch *img.Angle {
		case "front":
			resp.AngleImages.Front = img.URL
		case "back":
			resp.AngleImages.Back = img.URL
		case "side":
			resp.AngleImages.Side = img.URL
		case "top":
			resp.AngleImages.Top = img.URL
		}
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

type CategoryStockResponse struct {
	CategoryName string `json:"category_name"`
	ProductCount int64  `json:"product_count"`
	TotalStock   int64  `json:"total_stock"`
}

type InventoryReportResponse struct {
	TotalProducts     int64                   `json:"total_products"`
	TotalStockValue   int64                   `json:"total_stock_value"`
	LowStockCount     int64                   `json:"low_stock_count"`
	OutOfStockCount   int64                   `json:"out_of_stock_count"`
	CategoryBreakdown []CategoryStockResponse `json:"category_breakdown"`
}

func ToInventoryReportResponse(r *usecase.InventoryReportResult) InventoryReportResponse {
	resp := InventoryReportResponse{
		TotalProducts:   r.Summary.TotalProducts,
		TotalStockValue: r.Summary.TotalStockValue,
		LowStockCount:   r.Summary.LowStockCount,
		OutOfStockCount: r.Summary.OutOfStockCount,
	}
	for _, c := range r.Summary.CategoryBreakdown {
		resp.CategoryBreakdown = append(resp.CategoryBreakdown, CategoryStockResponse{
			CategoryName: c.CategoryName,
			ProductCount: c.ProductCount,
			TotalStock:   c.TotalStock,
		})
	}
	return resp
}
