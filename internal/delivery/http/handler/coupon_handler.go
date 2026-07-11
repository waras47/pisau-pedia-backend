package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type CouponHandler struct {
	usecase *usecase.CouponUsecase
}

func NewCouponHandler(usecase *usecase.CouponUsecase) *CouponHandler {
	return &CouponHandler{usecase: usecase}
}

func (h *CouponHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.usecase.ListCoupons(c.Request().Context(), page, perPage)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list coupons", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToCouponResponses(result.Coupons), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *CouponHandler) Create(c echo.Context) error {
	var req dto.CouponRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	coupon, err := h.usecase.CreateCoupon(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create coupon", nil)
	}
	return response.Success(c, http.StatusCreated, "Coupon created", dto.ToCouponResponse(coupon))
}

func (h *CouponHandler) Update(c echo.Context) error {
	var req dto.CouponRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	coupon, err := h.usecase.UpdateCoupon(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return response.Error(c, http.StatusNotFound, "coupon not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update coupon", nil)
	}
	return response.Success(c, http.StatusOK, "Coupon updated", dto.ToCouponResponse(coupon))
}

func (h *CouponHandler) Delete(c echo.Context) error {
	if err := h.usecase.DeleteCoupon(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return response.Error(c, http.StatusNotFound, "coupon not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete coupon", nil)
	}
	return response.Success(c, http.StatusOK, "Coupon deleted", nil)
}

func (h *CouponHandler) Validate(c echo.Context) error {
	var req dto.ValidateCouponRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	result, err := h.usecase.ValidateCoupon(c.Request().Context(), req.Code, req.Subtotal)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return response.Error(c, http.StatusNotFound, "coupon not found", nil)
		}
		if errors.Is(err, usecase.ErrCouponInvalid) {
			return response.Error(c, http.StatusUnprocessableEntity, "coupon is invalid or expired", nil)
		}
		if errors.Is(err, usecase.ErrCouponMinOrderNotMet) {
			return response.Error(c, http.StatusUnprocessableEntity, "order does not meet the coupon's minimum amount", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to validate coupon", nil)
	}
	return response.Success(c, http.StatusOK, "Coupon is valid", dto.ToValidateCouponResponse(result))
}
