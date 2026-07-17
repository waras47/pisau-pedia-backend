package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Phone    string `json:"phone" validate:"omitempty,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type GoogleExchangeRequest struct {
	Code string `json:"code" validate:"required"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
}

func ToUserResponse(u *entity.User) UserResponse {
	resp := UserResponse{
		ID:       u.ID,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     string(u.Role),
	}
	if u.Phone != nil {
		resp.Phone = *u.Phone
	}
	return resp
}

func ToAuthResponse(u *entity.User, tokens *usecase.TokenPair) AuthResponse {
	return AuthResponse{
		User:         ToUserResponse(u),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}
}
