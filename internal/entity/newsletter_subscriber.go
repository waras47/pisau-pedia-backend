package entity

import "time"

type SubscriberSource string

const (
	SubscriberSourceCheckout SubscriberSource = "checkout"
	SubscriberSourceFooter   SubscriberSource = "footer"
	SubscriberSourcePopup    SubscriberSource = "popup"
	SubscriberSourceBlog     SubscriberSource = "blog"
	SubscriberSourceManual   SubscriberSource = "manual"
)

type SubscriberStatus string

const (
	SubscriberStatusActive       SubscriberStatus = "active"
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
)

type NewsletterSubscriber struct {
	ID        string           `db:"id"`
	Email     string           `db:"email"`
	Name      *string          `db:"name"`
	Source    SubscriberSource `db:"source"`
	Status    SubscriberStatus `db:"status"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}
