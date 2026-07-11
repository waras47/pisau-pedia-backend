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

type NewsletterHandler struct {
	usecase *usecase.NewsletterUsecase
}

func NewNewsletterHandler(usecase *usecase.NewsletterUsecase) *NewsletterHandler {
	return &NewsletterHandler{usecase: usecase}
}

func (h *NewsletterHandler) Subscribe(c echo.Context) error {
	var req dto.SubscribeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	sub, err := h.usecase.Subscribe(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to subscribe", nil)
	}
	return response.Success(c, http.StatusCreated, "Subscribed", dto.ToSubscriberResponse(sub))
}

func (h *NewsletterHandler) ListSubscribers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.usecase.ListSubscribers(c.Request().Context(), usecase.SubscriberListInput{
		Page:    page,
		PerPage: perPage,
		Status:  c.QueryParam("status"),
		Search:  c.QueryParam("search"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list subscribers", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToSubscriberResponses(result.Subscribers), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *NewsletterHandler) Unsubscribe(c echo.Context) error {
	if err := h.usecase.Unsubscribe(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrSubscriberNotFound) {
			return response.Error(c, http.StatusNotFound, "subscriber not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to unsubscribe", nil)
	}
	return response.Success(c, http.StatusOK, "Unsubscribed", nil)
}

func (h *NewsletterHandler) Delete(c echo.Context) error {
	if err := h.usecase.DeleteSubscriber(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrSubscriberNotFound) {
			return response.Error(c, http.StatusNotFound, "subscriber not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete subscriber", nil)
	}
	return response.Success(c, http.StatusOK, "Subscriber deleted", nil)
}
