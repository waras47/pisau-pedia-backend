package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type SearchHandler struct {
	usecase *usecase.SearchUsecase
}

func NewSearchHandler(usecase *usecase.SearchUsecase) *SearchHandler {
	return &SearchHandler{usecase: usecase}
}

// Global powers the admin panel's quick-jump search — products, customers,
// and orders matching `q`, a handful of each, not a paginated results page.
func (h *SearchHandler) Global(c echo.Context) error {
	query := c.QueryParam("q")

	result, err := h.usecase.GlobalSearch(c.Request().Context(), query)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to search", nil)
	}

	return response.Success(c, http.StatusOK, "OK", dto.ToGlobalSearchResponse(result))
}
