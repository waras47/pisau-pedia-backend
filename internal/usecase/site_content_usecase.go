package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type SiteContentInput struct {
	Key   string
	Value string
}

type SiteContentUsecase struct {
	repo repository.SiteContentRepository
}

func NewSiteContentUsecase(repo repository.SiteContentRepository) *SiteContentUsecase {
	return &SiteContentUsecase{repo: repo}
}

func (u *SiteContentUsecase) List(ctx context.Context) ([]entity.SiteContent, error) {
	return u.repo.FindAll(ctx)
}

func (u *SiteContentUsecase) GetByKey(ctx context.Context, key string) (*entity.SiteContent, error) {
	return u.repo.FindByKey(ctx, key)
}

func (u *SiteContentUsecase) Upsert(ctx context.Context, input SiteContentInput) (*entity.SiteContent, error) {
	content := &entity.SiteContent{
		ID:    uuid.NewString(),
		Key:   input.Key,
		Value: input.Value,
	}
	if err := u.repo.Upsert(ctx, content); err != nil {
		return nil, err
	}
	return u.repo.FindByKey(ctx, input.Key)
}

func (u *SiteContentUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
