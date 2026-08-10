package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type PushHandler struct {
	repo          repository.PushSubscriptionRepository
	vapidPubKey   string
}

func NewPushHandler(repo repository.PushSubscriptionRepository, vapidPubKey string) *PushHandler {
	return &PushHandler{repo: repo, vapidPubKey: vapidPubKey}
}

func (h *PushHandler) GetVAPIDKey(c echo.Context) error {
	return response.Success(c, http.StatusOK, "OK", dto.VAPIDPublicKeyResponse{PublicKey: h.vapidPubKey})
}

func (h *PushHandler) Subscribe(c echo.Context) error {
	var req dto.PushSubscribeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error(), nil)
	}

	userID, _ := c.Get(middleware.ContextKeyUserID).(string)

	sub := &entity.PushSubscription{
		ID:       uuid.New().String(),
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256dh:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}

	if err := h.repo.Upsert(c.Request().Context(), sub); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to save subscription", nil)
	}
	return response.Success(c, http.StatusOK, "Subscribed", nil)
}

func (h *PushHandler) Unsubscribe(c echo.Context) error {
	var req dto.PushUnsubscribeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}

	if err := h.repo.DeleteByEndpoint(c.Request().Context(), req.Endpoint); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to unsubscribe", nil)
	}
	return response.Success(c, http.StatusOK, "Unsubscribed", nil)
}
