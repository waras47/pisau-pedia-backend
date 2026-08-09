package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type sitePromoRepository struct {
	db *sqlx.DB
}

func NewSitePromoRepository(db *sqlx.DB) repository.SitePromoRepository {
	return &sitePromoRepository{db: db}
}

func (r *sitePromoRepository) FindAll(ctx context.Context) ([]entity.SitePromo, error) {
	var promos []entity.SitePromo
	err := r.db.SelectContext(ctx, &promos, `SELECT * FROM site_promos ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	return promos, nil
}

func (r *sitePromoRepository) FindByID(ctx context.Context, id string) (*entity.SitePromo, error) {
	var promo entity.SitePromo
	err := r.db.GetContext(ctx, &promo, `SELECT * FROM site_promos WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("promo not found")
		}
		return nil, err
	}
	return &promo, nil
}

func (r *sitePromoRepository) FindActive(ctx context.Context) (*entity.SitePromo, error) {
	var promo entity.SitePromo
	err := r.db.GetContext(ctx, &promo, `
		SELECT * FROM site_promos
		WHERE is_active = 1 AND CURDATE() BETWEEN start_date AND end_date
		ORDER BY created_at DESC
		LIMIT 1
	`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &promo, nil
}

func (r *sitePromoRepository) Create(ctx context.Context, promo *entity.SitePromo) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO site_promos (id, title, description, discount_percent, popup_image, apply_to_all, start_date, end_date, is_active)
		VALUES (:id, :title, :description, :discount_percent, :popup_image, :apply_to_all, :start_date, :end_date, :is_active)
	`, promo)
	return err
}

func (r *sitePromoRepository) Update(ctx context.Context, promo *entity.SitePromo) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE site_promos SET
			title = :title,
			description = :description,
			discount_percent = :discount_percent,
			popup_image = :popup_image,
			apply_to_all = :apply_to_all,
			start_date = :start_date,
			end_date = :end_date,
			is_active = :is_active
		WHERE id = :id
	`, promo)
	return err
}

func (r *sitePromoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM site_promos WHERE id = ?`, id)
	return err
}

func (r *sitePromoRepository) SetProductIDs(ctx context.Context, promoID string, productIDs []string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM site_promo_products WHERE site_promo_id = ?`, promoID)
	if err != nil {
		return err
	}
	for _, pid := range productIDs {
		_, err = r.db.ExecContext(ctx, `INSERT INTO site_promo_products (site_promo_id, product_id) VALUES (?, ?)`, promoID, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *sitePromoRepository) GetProductIDs(ctx context.Context, promoID string) ([]string, error) {
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `SELECT product_id FROM site_promo_products WHERE site_promo_id = ?`, promoID)
	return ids, err
}
