package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type CollectionInput struct {
	Name        string
	Slug        string
	Description *string
	ImageURL    *string
	SortOrder   int
	CategoryIDs []string
}

type CollectionUsecase struct {
	collectionRepo repository.CollectionRepository
}

func NewCollectionUsecase(collectionRepo repository.CollectionRepository) *CollectionUsecase {
	return &CollectionUsecase{collectionRepo: collectionRepo}
}

func (u *CollectionUsecase) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	return u.collectionRepo.FindAll(ctx)
}

func (u *CollectionUsecase) GetCollectionBySlug(ctx context.Context, slug string) (*entity.Collection, error) {
	return u.collectionRepo.FindBySlug(ctx, slug)
}

func (u *CollectionUsecase) CreateCollection(ctx context.Context, input CollectionInput) (*entity.Collection, error) {
	base := input.Slug
	if base == "" {
		base = slugify(input.Name)
	} else {
		base = slugify(base)
	}

	slug, err := ensureUniqueSlug(ctx, base, u.collectionRepo.ExistsBySlug)
	if err != nil {
		return nil, err
	}

	collection := &entity.Collection{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		SortOrder:   input.SortOrder,
	}
	if err := u.collectionRepo.Create(ctx, collection); err != nil {
		return nil, err
	}

	if err := u.collectionRepo.SetCategories(ctx, collection.ID, input.CategoryIDs); err != nil {
		return nil, err
	}

	return u.collectionRepo.FindByID(ctx, collection.ID)
}

func (u *CollectionUsecase) UpdateCollection(ctx context.Context, id string, input CollectionInput) (*entity.Collection, error) {
	collection, err := u.collectionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		collection.Name = input.Name
	}
	if input.Description != nil {
		collection.Description = input.Description
	}
	if input.ImageURL != nil {
		collection.ImageURL = input.ImageURL
	}
	collection.SortOrder = input.SortOrder

	if err := u.collectionRepo.Update(ctx, collection); err != nil {
		return nil, err
	}

	if input.CategoryIDs != nil {
		if err := u.collectionRepo.SetCategories(ctx, id, input.CategoryIDs); err != nil {
			return nil, err
		}
	}

	return u.collectionRepo.FindByID(ctx, id)
}

func (u *CollectionUsecase) DeleteCollection(ctx context.Context, id string) error {
	if _, err := u.collectionRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.collectionRepo.Delete(ctx, id)
}
