package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ConfiguratorHandler struct {
	uc *usecase.ConfiguratorUsecase
}

func NewConfiguratorHandler(uc *usecase.ConfiguratorUsecase) *ConfiguratorHandler {
	return &ConfiguratorHandler{uc: uc}
}

// --- Shape ---

func (h *ConfiguratorHandler) ListShapes(c echo.Context) error {
	shapes, err := h.uc.ListShapes(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list shapes", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToShapeResponses(shapes))
}

func (h *ConfiguratorHandler) GetShape(c echo.Context) error {
	shape, err := h.uc.GetShape(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorShapeNotFound) {
			return response.Error(c, http.StatusNotFound, "shape not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get shape", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToShapeResponse(shape))
}

func (h *ConfiguratorHandler) CreateShape(c echo.Context) error {
	var req dto.CreateShapeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	shape, err := h.uc.CreateShape(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create shape", nil)
	}
	return response.Success(c, http.StatusCreated, "Shape created", dto.ToShapeResponse(shape))
}

func (h *ConfiguratorHandler) UpdateShape(c echo.Context) error {
	var req dto.UpdateShapeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	shape, err := h.uc.UpdateShape(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorShapeNotFound) {
			return response.Error(c, http.StatusNotFound, "shape not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update shape", nil)
	}
	return response.Success(c, http.StatusOK, "Shape updated", dto.ToShapeResponse(shape))
}

func (h *ConfiguratorHandler) DeleteShape(c echo.Context) error {
	if err := h.uc.DeleteShape(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrConfiguratorShapeNotFound) {
			return response.Error(c, http.StatusNotFound, "shape not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete shape", nil)
	}
	return response.Success(c, http.StatusOK, "Shape deleted", nil)
}

// --- Blade ---

func (h *ConfiguratorHandler) ListBlades(c echo.Context) error {
	blades, err := h.uc.ListBlades(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list blades", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToBladeResponses(blades))
}

func (h *ConfiguratorHandler) GetBlade(c echo.Context) error {
	blade, err := h.uc.GetBlade(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorBladeNotFound) {
			return response.Error(c, http.StatusNotFound, "blade not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get blade", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToBladeResponse(blade))
}

func (h *ConfiguratorHandler) CreateBlade(c echo.Context) error {
	var req dto.CreateBladeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	blade, err := h.uc.CreateBlade(c.Request().Context(), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorShapeNotFound) {
			return response.Error(c, http.StatusBadRequest, "shape not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create blade", nil)
	}
	return response.Success(c, http.StatusCreated, "Blade created", dto.ToBladeResponse(blade))
}

func (h *ConfiguratorHandler) UpdateBlade(c echo.Context) error {
	var req dto.UpdateBladeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	blade, err := h.uc.UpdateBlade(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorBladeNotFound) {
			return response.Error(c, http.StatusNotFound, "blade not found", nil)
		}
		if errors.Is(err, repository.ErrConfiguratorShapeNotFound) {
			return response.Error(c, http.StatusBadRequest, "shape not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update blade", nil)
	}
	return response.Success(c, http.StatusOK, "Blade updated", dto.ToBladeResponse(blade))
}

func (h *ConfiguratorHandler) DeleteBlade(c echo.Context) error {
	if err := h.uc.DeleteBlade(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrConfiguratorBladeNotFound) {
			return response.Error(c, http.StatusNotFound, "blade not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete blade", nil)
	}
	return response.Success(c, http.StatusOK, "Blade deleted", nil)
}

// --- Handle ---

func (h *ConfiguratorHandler) ListHandles(c echo.Context) error {
	handles, err := h.uc.ListHandles(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list handles", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToHandleResponses(handles))
}

func (h *ConfiguratorHandler) GetHandle(c echo.Context) error {
	handle, err := h.uc.GetHandle(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorHandleNotFound) {
			return response.Error(c, http.StatusNotFound, "handle not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get handle", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToHandleResponse(handle))
}

func (h *ConfiguratorHandler) CreateHandle(c echo.Context) error {
	var req dto.CreateHandleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	handle, err := h.uc.CreateHandle(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create handle", nil)
	}
	return response.Success(c, http.StatusCreated, "Handle created", dto.ToHandleResponse(handle))
}

func (h *ConfiguratorHandler) UpdateHandle(c echo.Context) error {
	var req dto.UpdateHandleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	handle, err := h.uc.UpdateHandle(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorHandleNotFound) {
			return response.Error(c, http.StatusNotFound, "handle not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update handle", nil)
	}
	return response.Success(c, http.StatusOK, "Handle updated", dto.ToHandleResponse(handle))
}

func (h *ConfiguratorHandler) DeleteHandle(c echo.Context) error {
	if err := h.uc.DeleteHandle(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrConfiguratorHandleNotFound) {
			return response.Error(c, http.StatusNotFound, "handle not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete handle", nil)
	}
	return response.Success(c, http.StatusOK, "Handle deleted", nil)
}

// --- Accessory ---

func (h *ConfiguratorHandler) ListAccessories(c echo.Context) error {
	accessories, err := h.uc.ListAccessories(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list accessories", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToAccessoryResponses(accessories))
}

func (h *ConfiguratorHandler) GetAccessory(c echo.Context) error {
	accessory, err := h.uc.GetAccessory(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorAccessoryNotFound) {
			return response.Error(c, http.StatusNotFound, "accessory not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get accessory", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToAccessoryResponse(accessory))
}

func (h *ConfiguratorHandler) CreateAccessory(c echo.Context) error {
	var req dto.CreateAccessoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	accessory, err := h.uc.CreateAccessory(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create accessory", nil)
	}
	return response.Success(c, http.StatusCreated, "Accessory created", dto.ToAccessoryResponse(accessory))
}

func (h *ConfiguratorHandler) UpdateAccessory(c echo.Context) error {
	var req dto.UpdateAccessoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	accessory, err := h.uc.UpdateAccessory(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrConfiguratorAccessoryNotFound) {
			return response.Error(c, http.StatusNotFound, "accessory not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update accessory", nil)
	}
	return response.Success(c, http.StatusOK, "Accessory updated", dto.ToAccessoryResponse(accessory))
}

func (h *ConfiguratorHandler) DeleteAccessory(c echo.Context) error {
	if err := h.uc.DeleteAccessory(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrConfiguratorAccessoryNotFound) {
			return response.Error(c, http.StatusNotFound, "accessory not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete accessory", nil)
	}
	return response.Success(c, http.StatusOK, "Accessory deleted", nil)
}
