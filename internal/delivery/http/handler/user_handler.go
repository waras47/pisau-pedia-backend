package handler

import (
	"errors"
	"net/http"
	"strconv"

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
		if errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
			return response.Error(c, http.StatusConflict, "email already registered", nil)
		}
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

func (h *UserHandler) ListCustomers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.userUsecase.ListCustomers(c.Request().Context(), usecase.CustomerListInput{
		Page:    page,
		PerPage: perPage,
		Search:  c.QueryParam("search"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list customers", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToCustomerResponses(result.Customers), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) GetCustomer(c echo.Context) error {
	customer, err := h.userUsecase.GetCustomer(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return response.Error(c, http.StatusNotFound, "customer not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get customer", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToCustomerResponse(customer))
}

func (h *UserHandler) GetCustomerAddresses(c echo.Context) error {
	addresses, err := h.userUsecase.ListCustomerAddresses(c.Request().Context(), c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list customer addresses", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToAddressResponses(addresses))
}

func (h *UserHandler) UpdateCustomerStatus(c echo.Context) error {
	var req dto.UpdateCustomerStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}

	user, err := h.userUsecase.UpdateCustomerStatus(c.Request().Context(), c.Param("id"), req.IsActive)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return response.Error(c, http.StatusNotFound, "customer not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update customer status", nil)
	}
	return response.Success(c, http.StatusOK, "Customer status updated", dto.ToUserResponse(user))
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
