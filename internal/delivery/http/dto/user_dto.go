package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type UpdateProfileRequest struct {
	FullName  *string `json:"full_name" validate:"omitempty,min=2"`
	Phone     *string `json:"phone" validate:"omitempty,min=8"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

func (r UpdateProfileRequest) ToPatch() usecase.ProfilePatch {
	return usecase.ProfilePatch{
		FullName:  r.FullName,
		Phone:     r.Phone,
		AvatarURL: r.AvatarURL,
	}
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type AddressRequest struct {
	Label       *string `json:"label"`
	FullName    string  `json:"full_name" validate:"required,min=2"`
	Phone       *string `json:"phone"`
	AddressLine string  `json:"address_line" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Province    *string `json:"province"`
	PostalCode  string  `json:"postal_code" validate:"required"`
	IsDefault   bool    `json:"is_default"`
}

func (r AddressRequest) ToInput() usecase.AddressInput {
	return usecase.AddressInput{
		Label:       r.Label,
		FullName:    r.FullName,
		Phone:       r.Phone,
		AddressLine: r.AddressLine,
		City:        r.City,
		Province:    r.Province,
		PostalCode:  r.PostalCode,
		IsDefault:   r.IsDefault,
	}
}

type AddressResponse struct {
	ID          string `json:"id"`
	Label       string `json:"label,omitempty"`
	FullName    string `json:"full_name"`
	Phone       string `json:"phone,omitempty"`
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	Province    string `json:"province,omitempty"`
	PostalCode  string `json:"postal_code"`
	IsDefault   bool   `json:"is_default"`
}

func ToAddressResponse(a *entity.Address) AddressResponse {
	resp := AddressResponse{
		ID:          a.ID,
		FullName:    a.FullName,
		AddressLine: a.AddressLine,
		City:        a.City,
		PostalCode:  a.PostalCode,
		IsDefault:   a.IsDefault,
	}
	if a.Label != nil {
		resp.Label = *a.Label
	}
	if a.Phone != nil {
		resp.Phone = *a.Phone
	}
	if a.Province != nil {
		resp.Province = *a.Province
	}
	return resp
}

func ToAddressResponses(addresses []entity.Address) []AddressResponse {
	out := make([]AddressResponse, 0, len(addresses))
	for i := range addresses {
		out = append(out, ToAddressResponse(&addresses[i]))
	}
	return out
}
