package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type SiteContentHandler struct {
	usecase *usecase.SiteContentUsecase
}

func NewSiteContentHandler(u *usecase.SiteContentUsecase) *SiteContentHandler {
	return &SiteContentHandler{usecase: u}
}

func (h *SiteContentHandler) List(c echo.Context) error {
	items, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list site contents", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToSiteContentResponses(items))
}

func (h *SiteContentHandler) GetByKey(c echo.Context) error {
	item, err := h.usecase.GetByKey(c.Request().Context(), c.Param("key"))
	if err != nil {
		return response.Error(c, http.StatusNotFound, "content not found", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToSiteContentResponse(item))
}

func (h *SiteContentHandler) Upsert(c echo.Context) error {
	var req dto.UpsertSiteContentRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	item, err := h.usecase.Upsert(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to save site content", nil)
	}
	return response.Success(c, http.StatusOK, "Content saved", dto.ToSiteContentResponse(item))
}

func (h *SiteContentHandler) Delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete site content", nil)
	}
	return response.Success(c, http.StatusOK, "Content deleted", nil)
}
