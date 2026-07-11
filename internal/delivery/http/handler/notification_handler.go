package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type NotificationHandler struct {
	usecase *usecase.NotificationUsecase
}

func NewNotificationHandler(usecase *usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: usecase}
}

func (h *NotificationHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	unreadOnly, _ := strconv.ParseBool(c.QueryParam("unread_only"))

	result, err := h.usecase.ListNotifications(c.Request().Context(), usecase.NotificationListInput{
		Page:       page,
		PerPage:    perPage,
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list notifications", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToNotificationResponses(result.Notifications), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *NotificationHandler) UnreadCount(c echo.Context) error {
	count, err := h.usecase.CountUnread(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to count unread notifications", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.UnreadCountResponse{Count: count})
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	if err := h.usecase.MarkAsRead(c.Request().Context(), c.Param("id")); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to mark notification as read", nil)
	}
	return response.Success(c, http.StatusOK, "Notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	if err := h.usecase.MarkAllAsRead(c.Request().Context()); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to mark all notifications as read", nil)
	}
	return response.Success(c, http.StatusOK, "All notifications marked as read", nil)
}
