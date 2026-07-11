package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type SubscriberFilter struct {
	Page    int
	PerPage int
	Status  string
	Search  string
}

type NewsletterRepository interface {
	FindAll(ctx context.Context, filter SubscriberFilter) ([]entity.NewsletterSubscriber, int64, error)
	FindByEmail(ctx context.Context, email string) (*entity.NewsletterSubscriber, error)
	FindByID(ctx context.Context, id string) (*entity.NewsletterSubscriber, error)
	Create(ctx context.Context, sub *entity.NewsletterSubscriber) error
	UpdateStatus(ctx context.Context, id string, status entity.SubscriberStatus) error
	Delete(ctx context.Context, id string) error
}
