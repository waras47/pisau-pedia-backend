package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateServiceRequestRequest struct {
	Type          string  `json:"type" validate:"required,oneof=sharpening engraving"`
	CustomerName  string  `json:"customer_name" validate:"required,min=2"`
	CustomerEmail string  `json:"customer_email" validate:"required,email"`
	CustomerPhone *string `json:"customer_phone"`
	Message       string  `json:"message" validate:"required,min=5"`
}

func (r CreateServiceRequestRequest) ToInput() usecase.CreateServiceRequestInput {
	return usecase.CreateServiceRequestInput{
		Type:          r.Type,
		CustomerName:  r.CustomerName,
		CustomerEmail: r.CustomerEmail,
		CustomerPhone: r.CustomerPhone,
		Message:       r.Message,
	}
}

type UpdateServiceRequestRequest struct {
	Status      string  `json:"status" validate:"omitempty,oneof=pending in_progress completed rejected"`
	QuotedPrice *int64  `json:"quoted_price" validate:"omitempty,min=0"`
	AdminNotes  *string `json:"admin_notes"`
}

func (r UpdateServiceRequestRequest) ToInput() usecase.UpdateServiceRequestInput {
	return usecase.UpdateServiceRequestInput{
		Status:      entity.ServiceRequestStatus(r.Status),
		QuotedPrice: r.QuotedPrice,
		AdminNotes:  r.AdminNotes,
	}
}

type ServiceRequestResponse struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone,omitempty"`
	Message       string `json:"message"`
	QuotedPrice   *int64 `json:"quoted_price,omitempty"`
	AdminNotes    string `json:"admin_notes,omitempty"`
	CreatedAt     string `json:"created_at"`
}

func ToServiceRequestResponse(r *entity.ServiceRequest) ServiceRequestResponse {
	resp := ServiceRequestResponse{
		ID:            r.ID,
		Type:          string(r.Type),
		Status:        string(r.Status),
		CustomerName:  r.CustomerName,
		CustomerEmail: r.CustomerEmail,
		Message:       r.Message,
		QuotedPrice:   r.QuotedPrice,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.CustomerPhone != nil {
		resp.CustomerPhone = *r.CustomerPhone
	}
	if r.AdminNotes != nil {
		resp.AdminNotes = *r.AdminNotes
	}
	return resp
}

func ToServiceRequestResponses(requests []entity.ServiceRequest) []ServiceRequestResponse {
	out := make([]ServiceRequestResponse, 0, len(requests))
	for i := range requests {
		out = append(out, ToServiceRequestResponse(&requests[i]))
	}
	return out
}
