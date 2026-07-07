package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type categoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.SelectContext(ctx, &categories, `SELECT * FROM categories ORDER BY name`)
	return categories, err
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var category entity.Category
	err := r.db.GetContext(ctx, &category, `SELECT * FROM categories WHERE slug = ?`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	var category entity.Category
	err := r.db.GetContext(ctx, &category, `SELECT * FROM categories WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM categories WHERE slug = ?`, slug)
	return count > 0, err
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, description, image_url)
		VALUES (:id, :name, :slug, :description, :image_url)
	`
	_, err := r.db.NamedExecContext(ctx, query, category)
	return err
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	query := `
		UPDATE categories SET
			name = :name,
			description = :description,
			image_url = :image_url
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, category)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	return err
}
