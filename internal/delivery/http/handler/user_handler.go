package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

func currentUserID(c echo.Context) string {
	userID, _ := c.Get(middleware.ContextKeyUserID).(string)
	return userID
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	user, err := h.userUsecase.GetProfile(c.Request().Context(), currentUserID(c))
	if err != nil {
		return response.Error(c, http.StatusNotFound, "user not found", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToUserResponse(user))
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, err := h.userUsecase.UpdateProfile(c.Request().Context(), currentUserID(c), req.ToPatch())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update profile", nil)
	}
	return response.Success(c, http.StatusOK, "Profile updated", dto.ToUserResponse(user))
}

func (h *UserHandler) ChangePassword(c echo.Context) error {
	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	err := h.userUsecase.ChangePassword(c.Request().Context(), currentUserID(c), req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCurrentPassword) {
			return response.Error(c, http.StatusUnauthorized, "current password is incorrect", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to change password", nil)
	}
	return response.Success(c, http.StatusOK, "Password changed", nil)
}

func (h *UserHandler) ListAddresses(c echo.Context) error {
	addresses, err := h.userUsecase.ListAddresses(c.Request().Context(), currentUserID(c))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list addresses", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToAddressResponses(addresses))
}

func (h *UserHandler) CreateAddress(c echo.Context) error {
	var req dto.AddressRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	address, err := h.userUsecase.CreateAddress(c.Request().Context(), currentUserID(c), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create address", nil)
	}
	return response.Success(c, http.StatusCreated, "Address created", dto.ToAddressResponse(address))
}

func (h *UserHandler) UpdateAddress(c echo.Context) error {
	var req dto.AddressRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	address, err := h.userUsecase.UpdateAddress(c.Request().Context(), currentUserID(c), c.Param("id"), req.ToInput())
	if err != nil {
		return addressErrorResponse(c, err)
	}
	return response.Success(c, http.StatusOK, "Address updated", dto.ToAddressResponse(address))
}

func (h *UserHandler) DeleteAddress(c echo.Context) error {
	err := h.userUsecase.DeleteAddress(c.Request().Context(), currentUserID(c), c.Param("id"))
	if err != nil {
		return addressErrorResponse(c, err)
	}
	return response.Success(c, http.StatusOK, "Address deleted", nil)
}

func addressErrorResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, usecase.ErrForbidden):
		return response.Error(c, http.StatusForbidden, "you do not have access to this address", nil)
	case errors.Is(err, repository.ErrAddressNotFound):
		return response.Error(c, http.StatusNotFound, "address not found", nil)
	default:
		return response.Error(c, http.StatusInternalServerError, "failed to process address", nil)
	}
}
