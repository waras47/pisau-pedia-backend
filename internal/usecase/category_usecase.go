package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type CategoryInput struct {
	Name        string
	Slug        string
	Description *string
	ImageURL    *string
}

type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUsecase(categoryRepo repository.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{categoryRepo: categoryRepo}
}

func (u *CategoryUsecase) ListCategories(ctx context.Context) ([]entity.Category, error) {
	return u.categoryRepo.FindAll(ctx)
}

func (u *CategoryUsecase) GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return u.categoryRepo.FindBySlug(ctx, slug)
}

func (u *CategoryUsecase) CreateCategory(ctx context.Context, input CategoryInput) (*entity.Category, error) {
	base := input.Slug
	if base == "" {
		base = slugify(input.Name)
	} else {
		base = slugify(base)
	}

	slug, err := ensureUniqueSlug(ctx, base, u.categoryRepo.ExistsBySlug)
	if err != nil {
		return nil, err
	}

	category := &entity.Category{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		ImageURL:    input.ImageURL,
	}
	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *CategoryUsecase) UpdateCategory(ctx context.Context, id string, input CategoryInput) (*entity.Category, error) {
	category, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Description != nil {
		category.Description = input.Description
	}
	if input.ImageURL != nil {
		category.ImageURL = input.ImageURL
	}

	if err := u.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *CategoryUsecase) DeleteCategory(ctx context.Context, id string) error {
	if _, err := u.categoryRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.categoryRepo.Delete(ctx, id)
}
