package entity

import "time"

type NotificationModule string

const (
	NotificationModuleProduct  NotificationModule = "product"
	NotificationModuleOrder    NotificationModule = "order"
	NotificationModuleCustomer NotificationModule = "customer"
	NotificationModuleService  NotificationModule = "service"
)

type Notification struct {
	ID          string             `db:"id"`
	Module      NotificationModule `db:"module"`
	Type        string             `db:"type"`
	Title       string             `db:"title"`
	Message     string             `db:"message"`
	ReferenceID *string            `db:"reference_id"`
	Link        *string            `db:"link"`
	IsRead      bool               `db:"is_read"`
	CreatedAt   time.Time          `db:"created_at"`
}
