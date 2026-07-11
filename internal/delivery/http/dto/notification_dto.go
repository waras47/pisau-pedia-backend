package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type NotificationResponse struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	ReferenceID string `json:"reference_id,omitempty"`
	Link        string `json:"link,omitempty"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

func ToNotificationResponse(n *entity.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID,
		Module:    string(n.Module),
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if n.ReferenceID != nil {
		resp.ReferenceID = *n.ReferenceID
	}
	if n.Link != nil {
		resp.Link = *n.Link
	}
	return resp
}

func ToNotificationResponses(notifications []entity.Notification) []NotificationResponse {
	out := make([]NotificationResponse, 0, len(notifications))
	for i := range notifications {
		out = append(out, ToNotificationResponse(&notifications[i]))
	}
	return out
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
