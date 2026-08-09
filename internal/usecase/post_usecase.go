package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

// ---------- PostCategory ----------

type PostCategoryInput struct {
	Slug   string
	NameID string
	NameEN string
	DescID *string
	DescEN *string
}

type PostCategoryUsecase struct {
	repo repository.PostCategoryRepository
}

func NewPostCategoryUsecase(repo repository.PostCategoryRepository) *PostCategoryUsecase {
	return &PostCategoryUsecase{repo: repo}
}

func (u *PostCategoryUsecase) List(ctx context.Context) ([]entity.PostCategory, error) {
	return u.repo.FindAll(ctx)
}

func (u *PostCategoryUsecase) GetBySlug(ctx context.Context, slug string) (*entity.PostCategory, error) {
	return u.repo.FindBySlug(ctx, slug)
}

func (u *PostCategoryUsecase) Create(ctx context.Context, input PostCategoryInput) (*entity.PostCategory, error) {
	cat := &entity.PostCategory{
		ID:     uuid.NewString(),
		Slug:   input.Slug,
		NameID: input.NameID,
		NameEN: input.NameEN,
		DescID: input.DescID,
		DescEN: input.DescEN,
	}
	if err := u.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (u *PostCategoryUsecase) Update(ctx context.Context, id string, input PostCategoryInput) (*entity.PostCategory, error) {
	cat, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cat.Slug = input.Slug
	cat.NameID = input.NameID
	cat.NameEN = input.NameEN
	cat.DescID = input.DescID
	cat.DescEN = input.DescEN
	if err := u.repo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (u *PostCategoryUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

// ---------- Post ----------

type PostInput struct {
	Slug           string
	CategoryID     string
	TitleID        string
	TitleEN        string
	ExcerptID      *string
	ExcerptEN      *string
	ContentID      *string
	ContentEN      *string
	Image          *string
	ReadingMinutes int
	Status         string
	PublishedAt    *string // ISO date string, parsed to time.Time
}

type PostUsecase struct {
	repo repository.PostRepository
}

func NewPostUsecase(repo repository.PostRepository) *PostUsecase {
	return &PostUsecase{repo: repo}
}

func (u *PostUsecase) List(ctx context.Context, filter repository.PostFilter) ([]entity.Post, int64, error) {
	return u.repo.FindAll(ctx, filter)
}

func (u *PostUsecase) GetBySlug(ctx context.Context, slug string) (*entity.Post, error) {
	return u.repo.FindBySlug(ctx, slug)
}

func (u *PostUsecase) Create(ctx context.Context, input PostInput) (*entity.Post, error) {
	post := &entity.Post{
		ID:             uuid.NewString(),
		Slug:           input.Slug,
		CategoryID:     input.CategoryID,
		TitleID:        input.TitleID,
		TitleEN:        input.TitleEN,
		ExcerptID:      input.ExcerptID,
		ExcerptEN:      input.ExcerptEN,
		ContentID:      input.ContentID,
		ContentEN:      input.ContentEN,
		Image:          input.Image,
		ReadingMinutes: input.ReadingMinutes,
		Status:         input.Status,
	}

	post.PublishedAt = parsePublishedAt(input.PublishedAt)
	if post.Status == "published" && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := u.repo.Create(ctx, post); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, post.ID)
}

func (u *PostUsecase) Update(ctx context.Context, id string, input PostInput) (*entity.Post, error) {
	post, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	post.Slug = input.Slug
	post.CategoryID = input.CategoryID
	post.TitleID = input.TitleID
	post.TitleEN = input.TitleEN
	post.ExcerptID = input.ExcerptID
	post.ExcerptEN = input.ExcerptEN
	post.ContentID = input.ContentID
	post.ContentEN = input.ContentEN
	post.Image = input.Image
	post.ReadingMinutes = input.ReadingMinutes
	post.Status = input.Status

	if input.PublishedAt != nil {
		post.PublishedAt = parsePublishedAt(input.PublishedAt)
	}
	if post.Status == "published" && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := u.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, post.ID)
}

func (u *PostUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func parsePublishedAt(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02T15:04:05Z", *s)
	if err != nil {
		t, err = time.Parse("2006-01-02", *s)
		if err != nil {
			return nil
		}
	}
	return &t
}
