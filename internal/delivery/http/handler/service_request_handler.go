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

type ServiceRequestHandler struct {
	usecase *usecase.ServiceRequestUsecase
}

func NewServiceRequestHandler(usecase *usecase.ServiceRequestUsecase) *ServiceRequestHandler {
	return &ServiceRequestHandler{usecase: usecase}
}

func (h *ServiceRequestHandler) Create(c echo.Context) error {
	var req dto.CreateServiceRequestRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	created, err := h.usecase.CreateRequest(c.Request().Context(), req.ToInput())
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidServiceRequestType) {
			return response.Error(c, http.StatusUnprocessableEntity, "invalid service request type", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create service request", nil)
	}
	return response.Success(c, http.StatusCreated, "Service request created", dto.ToServiceRequestResponse(created))
}

func (h *ServiceRequestHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.usecase.ListRequests(c.Request().Context(), usecase.ServiceRequestListInput{
		Page:    page,
		PerPage: perPage,
		Type:    c.QueryParam("type"),
		Status:  c.QueryParam("status"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list service requests", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToServiceRequestResponses(result.Requests), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *ServiceRequestHandler) GetByID(c echo.Context) error {
	req, err := h.usecase.GetRequest(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrServiceRequestNotFound) {
			return response.Error(c, http.StatusNotFound, "service request not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get service request", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToServiceRequestResponse(req))
}

func (h *ServiceRequestHandler) UpdateStatus(c echo.Context) error {
	var req dto.UpdateServiceRequestRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	updated, err := h.usecase.UpdateRequest(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrServiceRequestNotFound) {
			return response.Error(c, http.StatusNotFound, "service request not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update service request", nil)
	}
	return response.Success(c, http.StatusOK, "Service request updated", dto.ToServiceRequestResponse(updated))
}
