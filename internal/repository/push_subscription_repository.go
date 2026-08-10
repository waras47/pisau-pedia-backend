package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type PushSubscriptionRepository interface {
	Upsert(ctx context.Context, sub *entity.PushSubscription) error
	DeleteByEndpoint(ctx context.Context, endpoint string) error
	FindAllAdmin(ctx context.Context) ([]entity.PushSubscription, error)
}
