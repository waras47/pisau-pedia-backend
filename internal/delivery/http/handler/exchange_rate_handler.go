package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/exchangerate"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

// exchangeRateOverrideKey is the site_contents key an admin can set to pin
// the USD->IDR rate instead of following the live open.er-api.com rate.
const exchangeRateOverrideKey = "exchange_rate_override"

type exchangeRateOverride struct {
	Mode     string  `json:"mode"` // "auto" or "manual"
	UsdToIdr float64 `json:"usd_to_idr"`
}

type ExchangeRateHandler struct {
	client      *exchangerate.Client
	siteContent *usecase.SiteContentUsecase
}

func NewExchangeRateHandler(client *exchangerate.Client, siteContent *usecase.SiteContentUsecase) *ExchangeRateHandler {
	return &ExchangeRateHandler{client: client, siteContent: siteContent}
}

func (h *ExchangeRateHandler) Get(c echo.Context) error {
	snapshot, err := h.client.GetRates(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusServiceUnavailable, "failed to fetch exchange rates", nil)
	}

	// A manual override is optional — any failure to load or parse it just
	// means we keep the live rate, never fail the request over it.
	if content, err := h.siteContent.GetByKey(c.Request().Context(), exchangeRateOverrideKey); err == nil {
		var override exchangeRateOverride
		if err := json.Unmarshal([]byte(content.Value), &override); err == nil {
			if override.Mode == "manual" && override.UsdToIdr > 0 {
				ratesCopy := make(map[string]float64, len(snapshot.Rates))
				for k, v := range snapshot.Rates {
					ratesCopy[k] = v
				}
				ratesCopy["IDR"] = override.UsdToIdr
				snapshot = &exchangerate.Snapshot{
					Base:      snapshot.Base,
					Rates:     ratesCopy,
					UpdatedAt: snapshot.UpdatedAt,
				}
			}
		}
	}

	return response.Success(c, http.StatusOK, "OK", snapshot)
}
