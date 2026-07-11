package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/exchangerate"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ExchangeRateHandler struct {
	client *exchangerate.Client
}

func NewExchangeRateHandler(client *exchangerate.Client) *ExchangeRateHandler {
	return &ExchangeRateHandler{client: client}
}

func (h *ExchangeRateHandler) Get(c echo.Context) error {
	snapshot, err := h.client.GetRates(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusServiceUnavailable, "failed to fetch exchange rates", nil)
	}
	return response.Success(c, http.StatusOK, "OK", snapshot)
}
