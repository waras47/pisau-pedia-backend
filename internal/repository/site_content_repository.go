package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type SiteContentRepository interface {
	FindAll(ctx context.Context) ([]entity.SiteContent, error)
	FindByKey(ctx context.Context, key string) (*entity.SiteContent, error)
	Upsert(ctx context.Context, content *entity.SiteContent) error
	Delete(ctx context.Context, id string) error
}
