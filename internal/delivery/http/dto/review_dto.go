package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateReviewRequest struct {
	ProductSlug   string  `json:"product_slug" validate:"required"`
	CustomerName  string  `json:"customer_name" validate:"required,min=2"`
	CustomerEmail *string `json:"customer_email" validate:"omitempty,email"`
	Rating        uint    `json:"rating" validate:"required,min=1,max=5"`
	Content       string  `json:"content" validate:"required,min=5"`
}

func (r CreateReviewRequest) ToInput() usecase.CreateReviewInput {
	return usecase.CreateReviewInput{
		ProductSlug:   r.ProductSlug,
		CustomerName:  r.CustomerName,
		CustomerEmail: r.CustomerEmail,
		Rating:        r.Rating,
		Content:       r.Content,
	}
}

type UpdateReviewStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending approved rejected"`
}

type ReviewResponse struct {
	ID            string `json:"id"`
	ProductName   string `json:"product_name,omitempty"`
	ProductSlug   string `json:"product_slug,omitempty"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email,omitempty"`
	Rating        uint   `json:"rating"`
	Content       string `json:"content"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func ToReviewResponse(r *entity.Review) ReviewResponse {
	resp := ReviewResponse{
		ID:           r.ID,
		CustomerName: r.CustomerName,
		Rating:       r.Rating,
		Content:      r.Content,
		Status:       string(r.Status),
		CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.CustomerEmail != nil {
		resp.CustomerEmail = *r.CustomerEmail
	}
	if r.ProductName != nil {
		resp.ProductName = *r.ProductName
	}
	if r.ProductSlug != nil {
		resp.ProductSlug = *r.ProductSlug
	}
	return resp
}

func ToReviewResponses(reviews []entity.Review) []ReviewResponse {
	out := make([]ReviewResponse, 0, len(reviews))
	for i := range reviews {
		out = append(out, ToReviewResponse(&reviews[i]))
	}
	return out
}
