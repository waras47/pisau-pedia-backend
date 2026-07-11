package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type NotificationFilter struct {
	Page       int
	PerPage    int
	UnreadOnly bool
}

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	FindAll(ctx context.Context, filter NotificationFilter) ([]entity.Notification, int64, error)
	CountUnread(ctx context.Context) (int64, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context) error
}
