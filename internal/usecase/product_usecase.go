package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

const (
	defaultPerPage = 12
	maxPerPage     = 50
)

type ProductListInput struct {
	Page         int
	PerPage      int
	CategorySlug string
	Search       string
	Sort         string
}

type ProductListResult struct {
	Products   []entity.Product
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

type ProductSpecInput struct {
	Label string
	Value string
}

type ProductInput struct {
	CategoryID     *string
	Name           string
	Slug           string
	Description    *string
	Price          int64
	CompareAtPrice *int64
	Maker          *string
	Badge          *entity.Badge
	Stock          *uint
	IsActive       *bool
	Images         []string
	Specs          []ProductSpecInput
	Highlights     []string
}

type ProductUsecase struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

func NewProductUsecase(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) *ProductUsecase {
	return &ProductUsecase{productRepo: productRepo, categoryRepo: categoryRepo}
}

func (u *ProductUsecase) ListProducts(ctx context.Context, input ProductListInput) (*ProductListResult, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	filter := repository.ProductFilter{
		Page:         page,
		PerPage:      perPage,
		CategorySlug: input.CategorySlug,
		Search:       input.Search,
		Sort:         input.Sort,
	}

	products, total, err := u.productRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &ProductListResult{
		Products:   products,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (u *ProductUsecase) GetProductBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	product, err := u.productRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !product.IsActive {
		return nil, repository.ErrProductNotFound
	}
	return product, nil
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, input ProductInput) (*entity.Product, error) {
	base := input.Slug
	if base == "" {
		base = slugify(input.Name)
	} else {
		base = slugify(base)
	}

	slug, err := ensureUniqueSlug(ctx, base, u.productRepo.ExistsBySlug)
	if err != nil {
		return nil, err
	}

	product := &entity.Product{
		ID:             uuid.New().String(),
		CategoryID:     input.CategoryID,
		Name:           input.Name,
		Slug:           slug,
		Description:    input.Description,
		Price:          input.Price,
		CompareAtPrice: input.CompareAtPrice,
		Currency:       "IDR",
		Maker:          input.Maker,
		Badge:          input.Badge,
		IsActive:       true,
	}
	if input.Stock != nil {
		product.Stock = *input.Stock
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}

	for _, url := range input.Images {
		product.Images = append(product.Images, entity.ProductImage{ID: uuid.New().String(), URL: url})
	}
	for _, spec := range input.Specs {
		product.Specs = append(product.Specs, entity.ProductSpec{ID: uuid.New().String(), Label: spec.Label, Value: spec.Value})
	}
	for _, highlight := range input.Highlights {
		product.Highlights = append(product.Highlights, entity.ProductHighlight{ID: uuid.New().String(), Highlight: highlight})
	}

	if err := u.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (u *ProductUsecase) UpdateProduct(ctx context.Context, id string, input ProductInput) (*entity.Product, error) {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		product.Name = input.Name
	}
	if input.Description != nil {
		product.Description = input.Description
	}
	if input.Price != 0 {
		product.Price = input.Price
	}
	if input.CompareAtPrice != nil {
		product.CompareAtPrice = input.CompareAtPrice
	}
	if input.CategoryID != nil {
		product.CategoryID = input.CategoryID
	}
	if input.Maker != nil {
		product.Maker = input.Maker
	}
	if input.Badge != nil {
		product.Badge = input.Badge
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}
	if input.Stock != nil {
		product.Stock = *input.Stock
	}

	if err := u.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (u *ProductUsecase) DeleteProduct(ctx context.Context, id string) error {
	if _, err := u.productRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.productRepo.Delete(ctx, id)
}
