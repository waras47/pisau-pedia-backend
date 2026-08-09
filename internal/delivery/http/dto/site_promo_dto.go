package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateSitePromoRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     *string  `json:"description"`
	DiscountPercent int      `json:"discount_percent" validate:"required,min=1,max=100"`
	PopupImage      *string  `json:"popup_image"`
	ApplyToAll      *bool    `json:"apply_to_all"`
	ProductIDs      []string `json:"product_ids"`
	StartDate       string   `json:"start_date" validate:"required"`
	EndDate         string   `json:"end_date" validate:"required"`
	IsActive        *bool    `json:"is_active"`
}

func (r CreateSitePromoRequest) ToInput() usecase.SitePromoInput {
	return usecase.SitePromoInput{
		Title:           r.Title,
		Description:     r.Description,
		DiscountPercent: r.DiscountPercent,
		PopupImage:      r.PopupImage,
		ApplyToAll:      r.ApplyToAll,
		ProductIDs:      r.ProductIDs,
		StartDate:       r.StartDate,
		EndDate:         r.EndDate,
		IsActive:        r.IsActive,
	}
}

type UpdateSitePromoRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     *string  `json:"description"`
	DiscountPercent int      `json:"discount_percent" validate:"required,min=1,max=100"`
	PopupImage      *string  `json:"popup_image"`
	ApplyToAll      *bool    `json:"apply_to_all"`
	ProductIDs      []string `json:"product_ids"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	IsActive        *bool    `json:"is_active"`
}

func (r UpdateSitePromoRequest) ToInput() usecase.SitePromoInput {
	return usecase.SitePromoInput{
		Title:           r.Title,
		Description:     r.Description,
		DiscountPercent: r.DiscountPercent,
		PopupImage:      r.PopupImage,
		ApplyToAll:      r.ApplyToAll,
		ProductIDs:      r.ProductIDs,
		StartDate:       r.StartDate,
		EndDate:         r.EndDate,
		IsActive:        r.IsActive,
	}
}

type SitePromoResponse struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	DiscountPercent int      `json:"discount_percent"`
	PopupImage      *string  `json:"popup_image,omitempty"`
	ApplyToAll      bool     `json:"apply_to_all"`
	ProductIDs      []string `json:"product_ids"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	IsActive        bool     `json:"is_active"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

func ToSitePromoResponse(p *entity.SitePromo) SitePromoResponse {
	productIDs := p.ProductIDs
	if productIDs == nil {
		productIDs = []string{}
	}
	return SitePromoResponse{
		ID:              p.ID,
		Title:           p.Title,
		Description:     p.Description,
		DiscountPercent: p.DiscountPercent,
		PopupImage:      p.PopupImage,
		ApplyToAll:      p.ApplyToAll,
		ProductIDs:      productIDs,
		StartDate:       p.StartDate.Format("2006-01-02"),
		EndDate:         p.EndDate.Format("2006-01-02"),
		IsActive:        p.IsActive,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToSitePromoListResponse(promos []entity.SitePromo) []SitePromoResponse {
	out := make([]SitePromoResponse, 0, len(promos))
	for i := range promos {
		out = append(out, ToSitePromoResponse(&promos[i]))
	}
	return out
}
