package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

// These are deliberately narrower than ProductDetailResponse/OrderResponse
// — the global search dropdown only needs enough to identify a result and
// link to it, not the full resource payload.

type SearchProductResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Price int64  `json:"price"`
	Image string `json:"image,omitempty"`
}

type SearchCustomerResult struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type SearchOrderResult struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	Total         int64  `json:"total"`
	CreatedAt     string `json:"created_at"`
}

type GlobalSearchResponse struct {
	Products  []SearchProductResult  `json:"products"`
	Customers []SearchCustomerResult `json:"customers"`
	Orders    []SearchOrderResult    `json:"orders"`
}

func ToGlobalSearchResponse(r *usecase.GlobalSearchResult) GlobalSearchResponse {
	resp := GlobalSearchResponse{
		Products:  make([]SearchProductResult, 0, len(r.Products)),
		Customers: make([]SearchCustomerResult, 0, len(r.Customers)),
		Orders:    make([]SearchOrderResult, 0, len(r.Orders)),
	}

	for _, p := range r.Products {
		item := SearchProductResult{ID: p.ID, Name: p.Name, Slug: p.Slug, Price: p.Price}
		if p.Image != nil {
			item.Image = *p.Image
		}
		resp.Products = append(resp.Products, item)
	}

	for _, c := range r.Customers {
		resp.Customers = append(resp.Customers, SearchCustomerResult{
			ID: c.ID, FullName: c.FullName, Email: c.Email,
		})
	}

	for _, o := range r.Orders {
		resp.Orders = append(resp.Orders, SearchOrderResult{
			ID:            o.ID,
			Status:        string(o.Status),
			CustomerName:  o.CustomerName,
			CustomerEmail: o.CustomerEmail,
			Total:         o.Total,
			CreatedAt:     o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return resp
}
