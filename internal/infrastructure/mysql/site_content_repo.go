package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type siteContentRepository struct {
	db *sqlx.DB
}

func NewSiteContentRepository(db *sqlx.DB) repository.SiteContentRepository {
	return &siteContentRepository{db: db}
}

func (r *siteContentRepository) FindAll(ctx context.Context) ([]entity.SiteContent, error) {
	var contents []entity.SiteContent
	err := r.db.SelectContext(ctx, &contents, "SELECT * FROM site_contents ORDER BY `key` ASC")
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *siteContentRepository) FindByKey(ctx context.Context, key string) (*entity.SiteContent, error) {
	var content entity.SiteContent
	err := r.db.GetContext(ctx, &content, "SELECT * FROM site_contents WHERE `key` = ?", key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}
	return &content, nil
}

func (r *siteContentRepository) Upsert(ctx context.Context, content *entity.SiteContent) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO site_contents (id, `+"`key`"+`, value)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value)
	`, content.ID, content.Key, content.Value)
	return err
}

func (r *siteContentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM site_contents WHERE id = ?", id)
	return err
}
