package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, tokens, err := h.authUsecase.Register(c.Request().Context(), req.Email, req.Password, req.FullName, req.Phone)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
			return response.Error(c, http.StatusConflict, "email already registered", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to register", nil)
	}

	return response.Success(c, http.StatusCreated, "Registration successful", dto.ToAuthResponse(user, tokens))
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, tokens, err := h.authUsecase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		case errors.Is(err, usecase.ErrAccountDisabled):
			return response.Error(c, http.StatusForbidden, "account disabled", nil)
		default:
			return response.Error(c, http.StatusInternalServerError, "failed to login", nil)
		}
	}

	return response.Success(c, http.StatusOK, "Login successful", dto.ToAuthResponse(user, tokens))
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	tokens, err := h.authUsecase.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token", nil)
	}

	return response.Success(c, http.StatusOK, "Token refreshed", map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req dto.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	if err := h.authUsecase.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to logout", nil)
	}

	return response.Success(c, http.StatusOK, "Logged out", nil)
}
