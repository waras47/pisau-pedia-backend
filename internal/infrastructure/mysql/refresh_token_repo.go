package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (:id, :user_id, :token_hash, :expires_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.GetContext(ctx, &token, `SELECT * FROM refresh_tokens WHERE token_hash = ?`, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, time.Now(), userID)
	return err
}
