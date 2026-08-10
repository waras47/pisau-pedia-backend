package entity

import "time"

type PushSubscription struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Endpoint  string    `db:"endpoint"`
	P256dh    string    `db:"p256dh"`
	Auth      string    `db:"auth"`
	CreatedAt time.Time `db:"created_at"`
}
