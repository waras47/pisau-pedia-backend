package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type PaymentHandler struct {
	usecase *usecase.PaymentUsecase
}

func NewPaymentHandler(usecase *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{usecase: usecase}
}

func (h *PaymentHandler) Methods(c echo.Context) error {
	methods, err := h.usecase.GetMethods(c.Request().Context())
	if err != nil {
		if errors.Is(err, usecase.ErrPaymentUnavailable) {
			return response.Success(c, http.StatusOK, "OK", []dto.PaymentMethodResponse{})
		}
		return response.Error(c, http.StatusInternalServerError, "failed to fetch payment methods", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToPaymentMethodResponses(methods))
}

// CheckStatus is the admin-triggered fallback for when the Komerce webhook
// never arrives (unreachable localhost during dev, or a missed delivery in
// production) — it asks Komerce directly instead of waiting.
func (h *PaymentHandler) CheckStatus(c echo.Context) error {
	id := c.Param("id")
	order, err := h.usecase.CheckStatus(c.Request().Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPaymentUnavailable):
			return response.Error(c, http.StatusServiceUnavailable, "payment service is not configured", nil)
		case errors.Is(err, usecase.ErrNoPaymentToCheck):
			return response.Error(c, http.StatusUnprocessableEntity, "order has no komerce payment to check", nil)
		case errors.Is(err, repository.ErrOrderNotFound):
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		default:
			return response.Error(c, http.StatusInternalServerError, "failed to check payment status", nil)
		}
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToOrderResponse(order))
}

// KomerceCallback is the webhook Komerce calls on payment status change. It
// must read the RAW body to verify the HMAC signature before parsing.
func (h *PaymentHandler) KomerceCallback(c echo.Context) error {
	rawBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "cannot read body", nil)
	}
	signature := c.Request().Header.Get("X-Callback-Api-Key")

	if err := h.usecase.HandleKomerceCallback(c.Request().Context(), rawBody, signature); err != nil {
		if errors.Is(err, usecase.ErrInvalidCallbackSignature) {
			return response.Error(c, http.StatusUnauthorized, "invalid signature", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to process callback", nil)
	}
	return response.Success(c, http.StatusOK, "OK", nil)
}
