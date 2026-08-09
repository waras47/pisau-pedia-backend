package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

// ---------- PostCategory ----------

type postCategoryRepository struct {
	db *sqlx.DB
}

func NewPostCategoryRepository(db *sqlx.DB) repository.PostCategoryRepository {
	return &postCategoryRepository{db: db}
}

func (r *postCategoryRepository) FindAll(ctx context.Context) ([]entity.PostCategory, error) {
	var cats []entity.PostCategory
	err := r.db.SelectContext(ctx, &cats, `SELECT * FROM post_categories ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *postCategoryRepository) FindByID(ctx context.Context, id string) (*entity.PostCategory, error) {
	var cat entity.PostCategory
	err := r.db.GetContext(ctx, &cat, `SELECT * FROM post_categories WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post category not found")
		}
		return nil, err
	}
	return &cat, nil
}

func (r *postCategoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.PostCategory, error) {
	var cat entity.PostCategory
	err := r.db.GetContext(ctx, &cat, `SELECT * FROM post_categories WHERE slug = ?`, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post category not found")
		}
		return nil, err
	}
	return &cat, nil
}

func (r *postCategoryRepository) Create(ctx context.Context, cat *entity.PostCategory) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO post_categories (id, slug, name_id, name_en, desc_id, desc_en)
		VALUES (:id, :slug, :name_id, :name_en, :desc_id, :desc_en)
	`, cat)
	return err
}

func (r *postCategoryRepository) Update(ctx context.Context, cat *entity.PostCategory) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE post_categories SET
			slug = :slug,
			name_id = :name_id,
			name_en = :name_en,
			desc_id = :desc_id,
			desc_en = :desc_en
		WHERE id = :id
	`, cat)
	return err
}

func (r *postCategoryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM post_categories WHERE id = ?`, id)
	return err
}

// ---------- Post ----------

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) repository.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) FindAll(ctx context.Context, filter repository.PostFilter) ([]entity.Post, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, filter.Status)
	}
	if filter.CategorySlug != "" {
		conditions = append(conditions, "pc.slug = ?")
		args = append(args, filter.CategorySlug)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderDir := "DESC"
	if filter.Sort == "oldest" {
		orderDir = "ASC"
	}

	// Count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM posts p
		LEFT JOIN post_categories pc ON pc.id = p.category_id
		%s
	`, where)

	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Paginate
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(`
		SELECT p.*, pc.slug AS category_slug, pc.name_id AS category_name
		FROM posts p
		LEFT JOIN post_categories pc ON pc.id = p.category_id
		%s
		ORDER BY p.published_at %s, p.created_at %s
		LIMIT ? OFFSET ?
	`, where, orderDir, orderDir)

	dataArgs := append(args, perPage, offset)
	var posts []entity.Post
	err = r.db.SelectContext(ctx, &posts, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) FindByID(ctx context.Context, id string) (*entity.Post, error) {
	var post entity.Post
	err := r.db.GetContext(ctx, &post, `
		SELECT p.*, pc.slug AS category_slug, pc.name_id AS category_name
		FROM posts p
		LEFT JOIN post_categories pc ON pc.id = p.category_id
		WHERE p.id = ?
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) FindBySlug(ctx context.Context, slug string) (*entity.Post, error) {
	var post entity.Post
	err := r.db.GetContext(ctx, &post, `
		SELECT p.*, pc.slug AS category_slug, pc.name_id AS category_name
		FROM posts p
		LEFT JOIN post_categories pc ON pc.id = p.category_id
		WHERE p.slug = ?
	`, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Create(ctx context.Context, post *entity.Post) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO posts (id, slug, category_id, title_id, title_en, excerpt_id, excerpt_en, content_id, content_en, image, reading_minutes, status, published_at)
		VALUES (:id, :slug, :category_id, :title_id, :title_en, :excerpt_id, :excerpt_en, :content_id, :content_en, :image, :reading_minutes, :status, :published_at)
	`, post)
	return err
}

func (r *postRepository) Update(ctx context.Context, post *entity.Post) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE posts SET
			slug = :slug,
			category_id = :category_id,
			title_id = :title_id,
			title_en = :title_en,
			excerpt_id = :excerpt_id,
			excerpt_en = :excerpt_en,
			content_id = :content_id,
			content_en = :content_en,
			image = :image,
			reading_minutes = :reading_minutes,
			status = :status,
			published_at = :published_at
		WHERE id = :id
	`, post)
	return err
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}
