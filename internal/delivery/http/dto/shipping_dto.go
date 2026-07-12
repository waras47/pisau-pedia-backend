package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/rajaongkir"
)

type DestinationResponse struct {
	ID              int64  `json:"id"`
	Label           string `json:"label"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	ZipCode         string `json:"zip_code"`
}

func ToDestinationResponses(dests []rajaongkir.Destination) []DestinationResponse {
	out := make([]DestinationResponse, 0, len(dests))
	for _, d := range dests {
		out = append(out, DestinationResponse{
			ID:              d.ID,
			Label:           d.Label,
			ProvinceName:    d.ProvinceName,
			CityName:        d.CityName,
			DistrictName:    d.DistrictName,
			SubdistrictName: d.SubdistrictName,
			ZipCode:         d.ZipCode,
		})
	}
	return out
}

type ShippingCostItemRequest struct {
	ProductSlug string `json:"product_slug" validate:"required"`
	Quantity    uint   `json:"quantity" validate:"required,min=1"`
}

type ShippingCostRequest struct {
	DestinationID string                    `json:"destination_id" validate:"required"`
	Items         []ShippingCostItemRequest `json:"items" validate:"required,min=1,dive"`
}

func (r ShippingCostRequest) ToItemInputs() []usecase.OrderItemInput {
	items := make([]usecase.OrderItemInput, 0, len(r.Items))
	for _, i := range r.Items {
		items = append(items, usecase.OrderItemInput{ProductSlug: i.ProductSlug, Quantity: i.Quantity})
	}
	return items
}

type ShippingOptionResponse struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	ETD         string `json:"etd"`
}

func ToShippingOptionResponses(opts []rajaongkir.ShippingOption) []ShippingOptionResponse {
	out := make([]ShippingOptionResponse, 0, len(opts))
	for _, o := range opts {
		out = append(out, ShippingOptionResponse{
			Name:        o.Name,
			Code:        o.Code,
			Service:     o.Service,
			Description: o.Description,
			Cost:        o.Cost,
			ETD:         o.ETD,
		})
	}
	return out
}
