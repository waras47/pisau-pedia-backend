package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type PostCategoryFilter struct {
	Page    int
	PerPage int
}

type PostFilter struct {
	Page         int
	PerPage      int
	Status       string
	CategorySlug string
	Sort         string // "newest" (default) or "oldest"
}

type PostCategoryRepository interface {
	FindAll(ctx context.Context) ([]entity.PostCategory, error)
	FindByID(ctx context.Context, id string) (*entity.PostCategory, error)
	FindBySlug(ctx context.Context, slug string) (*entity.PostCategory, error)
	Create(ctx context.Context, cat *entity.PostCategory) error
	Update(ctx context.Context, cat *entity.PostCategory) error
	Delete(ctx context.Context, id string) error
}

type PostRepository interface {
	FindAll(ctx context.Context, filter PostFilter) ([]entity.Post, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Post, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Post, error)
	Create(ctx context.Context, post *entity.Post) error
	Update(ctx context.Context, post *entity.Post) error
	Delete(ctx context.Context, id string) error
}
