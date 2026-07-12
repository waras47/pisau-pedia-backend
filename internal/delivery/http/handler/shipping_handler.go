package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ShippingHandler struct {
	usecase *usecase.ShippingUsecase
}

func NewShippingHandler(usecase *usecase.ShippingUsecase) *ShippingHandler {
	return &ShippingHandler{usecase: usecase}
}

func (h *ShippingHandler) SearchDestinations(c echo.Context) error {
	search := c.QueryParam("search")
	if len(search) < 3 {
		return response.Success(c, http.StatusOK, "OK", []dto.DestinationResponse{})
	}

	dests, err := h.usecase.SearchDestinations(c.Request().Context(), search)
	if err != nil {
		if errors.Is(err, usecase.ErrShippingUnavailable) {
			return response.Error(c, http.StatusServiceUnavailable, "shipping service is not configured", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to search destinations", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToDestinationResponses(dests))
}

func (h *ShippingHandler) CalculateCost(c echo.Context) error {
	var req dto.ShippingCostRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	opts, err := h.usecase.CalculateOptions(c.Request().Context(), req.DestinationID, req.ToItemInputs())
	if err != nil {
		if errors.Is(err, usecase.ErrShippingUnavailable) {
			return response.Error(c, http.StatusServiceUnavailable, "shipping service is not configured", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to calculate shipping cost", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToShippingOptionResponses(opts))
}
