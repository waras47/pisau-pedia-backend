package entity

import "time"

type ServiceRequestType string

const (
	ServiceRequestTypeSharpening ServiceRequestType = "sharpening"
	ServiceRequestTypeEngraving  ServiceRequestType = "engraving"
)

type ServiceRequestStatus string

const (
	ServiceRequestStatusPending    ServiceRequestStatus = "pending"
	ServiceRequestStatusInProgress ServiceRequestStatus = "in_progress"
	ServiceRequestStatusCompleted  ServiceRequestStatus = "completed"
	ServiceRequestStatusRejected   ServiceRequestStatus = "rejected"
)

type ServiceRequest struct {
	ID            string               `db:"id"`
	Type          ServiceRequestType   `db:"type"`
	Status        ServiceRequestStatus `db:"status"`
	CustomerName  string               `db:"customer_name"`
	CustomerEmail string               `db:"customer_email"`
	CustomerPhone *string              `db:"customer_phone"`
	Message       string               `db:"message"`
	QuotedPrice   *int64               `db:"quoted_price"`
	AdminNotes    *string              `db:"admin_notes"`
	CreatedAt     time.Time            `db:"created_at"`
	UpdatedAt     time.Time            `db:"updated_at"`
}
