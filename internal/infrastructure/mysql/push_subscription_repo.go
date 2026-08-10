package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type PushSubscriptionRepository struct {
	db *sqlx.DB
}

func NewPushSubscriptionRepository(db *sqlx.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

func (r *PushSubscriptionRepository) Upsert(ctx context.Context, sub *entity.PushSubscription) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE user_id = VALUES(user_id), p256dh = VALUES(p256dh), auth = VALUES(auth)`,
		sub.ID, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth,
	)
	return err
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE endpoint = ?`, endpoint)
	return err
}

func (r *PushSubscriptionRepository) FindAllAdmin(ctx context.Context) ([]entity.PushSubscription, error) {
	var subs []entity.PushSubscription
	err := r.db.SelectContext(ctx,
		&subs,
		`SELECT ps.* FROM push_subscriptions ps
		 JOIN users u ON u.id = ps.user_id
		 WHERE u.role = 'admin'`)
	return subs, err
}
