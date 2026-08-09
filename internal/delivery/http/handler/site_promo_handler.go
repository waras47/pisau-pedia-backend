package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type SitePromoHandler struct {
	usecase *usecase.SitePromoUsecase
}

func NewSitePromoHandler(u *usecase.SitePromoUsecase) *SitePromoHandler {
	return &SitePromoHandler{usecase: u}
}

func (h *SitePromoHandler) List(c echo.Context) error {
	promos, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list promos", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToSitePromoListResponse(promos))
}

func (h *SitePromoHandler) GetActive(c echo.Context) error {
	promo, err := h.usecase.GetActive(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to get active promo", nil)
	}
	if promo == nil {
		return response.Success(c, http.StatusOK, "no active promo", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToSitePromoResponse(promo))
}

func (h *SitePromoHandler) Create(c echo.Context) error {
	var req dto.CreateSitePromoRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	promo, err := h.usecase.Create(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create promo", nil)
	}
	return response.Success(c, http.StatusCreated, "Promo created", dto.ToSitePromoResponse(promo))
}

func (h *SitePromoHandler) Update(c echo.Context) error {
	var req dto.UpdateSitePromoRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	promo, err := h.usecase.Update(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update promo", nil)
	}
	return response.Success(c, http.StatusOK, "Promo updated", dto.ToSitePromoResponse(promo))
}

func (h *SitePromoHandler) Delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete promo", nil)
	}
	return response.Success(c, http.StatusOK, "Promo deleted", nil)
}

func (h *SitePromoHandler) UploadImage(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "use /admin/uploads/image endpoint"})
}
