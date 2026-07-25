package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type couponRepository struct {
	db *sqlx.DB
}

func NewCouponRepository(db *sqlx.DB) repository.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) FindAll(ctx context.Context, filter repository.CouponFilter) ([]entity.Coupon, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM coupons`); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	var coupons []entity.Coupon
	if err := r.db.SelectContext(ctx, &coupons, `
		SELECT * FROM coupons ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, filter.PerPage, offset); err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

func (r *couponRepository) FindByID(ctx context.Context, id string) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.GetContext(ctx, &coupon, `SELECT * FROM coupons WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrCouponNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) FindByCode(ctx context.Context, code string) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.GetContext(ctx, &coupon, `SELECT * FROM coupons WHERE code = ?`, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrCouponNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) FindPopup(ctx context.Context) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.GetContext(ctx, &coupon, `
		SELECT * FROM coupons
		WHERE show_popup = TRUE AND is_active = TRUE
			AND (starts_at IS NULL OR starts_at <= NOW())
			AND (ends_at IS NULL OR ends_at >= NOW())
			AND (max_uses IS NULL OR used_count < max_uses)
		LIMIT 1
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrCouponNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) Create(ctx context.Context, coupon *entity.Coupon) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO coupons (id, code, type, value, min_order, max_uses, starts_at, ends_at, is_active, show_popup, description)
		VALUES (:id, :code, :type, :value, :min_order, :max_uses, :starts_at, :ends_at, :is_active, :show_popup, :description)
	`, coupon)
	return err
}

func (r *couponRepository) Update(ctx context.Context, coupon *entity.Coupon) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE coupons SET
			code = :code, type = :type, value = :value, min_order = :min_order,
			max_uses = :max_uses, starts_at = :starts_at, ends_at = :ends_at,
			is_active = :is_active, show_popup = :show_popup, description = :description
		WHERE id = :id
	`, coupon)
	return err
}

func (r *couponRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id = ?`, id)
	return err
}

func (r *couponRepository) IncrementUsage(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE coupons SET used_count = used_count + 1 WHERE id = ?`, id)
	return err
}
