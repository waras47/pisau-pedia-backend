package mysql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO notifications (id, module, type, title, message, reference_id, link, is_read)
		VALUES (:id, :module, :type, :title, :message, :reference_id, :link, :is_read)
	`, n)
	return err
}

func (r *notificationRepository) FindAll(ctx context.Context, filter repository.NotificationFilter) ([]entity.Notification, int64, error) {
	where := "1=1"
	if filter.UnreadOnly {
		where = "is_read = 0"
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM notifications WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT * FROM notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	var notifications []entity.Notification
	if err := r.db.SelectContext(ctx, &notifications, listQuery, filter.PerPage, offset); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) CountUnread(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM notifications WHERE is_read = 0`)
	return count, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = 1 WHERE id = ?`, id)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = 1 WHERE is_read = 0`)
	return err
}
