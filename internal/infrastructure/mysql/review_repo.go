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

type reviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) FindAll(ctx context.Context, filter repository.ReviewFilter) ([]entity.Review, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Status != "" {
		conditions = append(conditions, "r.status = ?")
		args = append(args, filter.Status)
	}
	if filter.ProductSlug != "" {
		conditions = append(conditions, "p.slug = ?")
		args = append(args, filter.ProductSlug)
	}
	switch filter.Scope {
	case "product":
		conditions = append(conditions, "r.product_id IS NOT NULL")
	case "shop":
		conditions = append(conditions, "r.product_id IS NULL")
	}

	where := "1=1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	// LEFT JOIN (not JOIN) so shop reviews — which have no product_id —
	// still come back, just with product_name/product_slug left NULL.
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM reviews r
		LEFT JOIN products p ON p.id = r.product_id
		WHERE %s
	`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT r.*, p.name AS product_name, p.slug AS product_slug
		FROM reviews r
		LEFT JOIN products p ON p.id = r.product_id
		WHERE %s
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var reviews []entity.Review
	if err := r.db.SelectContext(ctx, &reviews, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id string) (*entity.Review, error) {
	var review entity.Review
	query := `
		SELECT r.*, p.name AS product_name, p.slug AS product_slug
		FROM reviews r
		LEFT JOIN products p ON p.id = r.product_id
		WHERE r.id = ?
	`
	if err := r.db.GetContext(ctx, &review, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReviewNotFound
		}
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) Create(ctx context.Context, review *entity.Review) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO reviews (id, product_id, customer_name, customer_email, rating, content, photos, status)
		VALUES (:id, :product_id, :customer_name, :customer_email, :rating, :content, :photos, :status)
	`, review)
	return err
}

func (r *reviewRepository) Update(ctx context.Context, review *entity.Review) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reviews
		SET customer_name = ?, customer_email = ?, rating = ?, content = ?, photos = ?, status = ?
		WHERE id = ?
	`, review.CustomerName, review.CustomerEmail, review.Rating, review.Content, review.Photos, review.Status, review.ID)
	return err
}

func (r *reviewRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM reviews WHERE id = ?`, id)
	return err
}

func (r *reviewRepository) GetApprovedStatsByProduct(ctx context.Context, productID string) (*repository.ProductRatingStats, error) {
	var stats struct {
		RatingAvg   float64 `db:"rating_avg"`
		ReviewCount uint    `db:"review_count"`
	}
	err := r.db.GetContext(ctx, &stats, `
		SELECT COALESCE(AVG(rating), 0) AS rating_avg, COUNT(*) AS review_count
		FROM reviews
		WHERE product_id = ? AND status = 'approved'
	`, productID)
	if err != nil {
		return nil, err
	}
	return &repository.ProductRatingStats{RatingAvg: stats.RatingAvg, ReviewCount: stats.ReviewCount}, nil
}
