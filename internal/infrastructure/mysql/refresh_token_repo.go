package mysql

import (
	"context"
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
	query :=
		`INSERT INTO refresh_token(
			id, 
			user_id, 
			token_hash, 
			expires_at
		) 
		VALUES(
			:id, 
			:user_id, 
			:token_hash, 
			:expires_at
		)
		`
	_, err := r.db.NamedExecContext(ctx, query, token)

	return err
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	query :=
		`SELECT 
			id, 
			user_id, 
			token_hash, 
			expires_at, 
			revoked_at, 
			created_at, 
			updated_at 
	    FROM refresh_token 
		WHERE token_hash = ?
	   `

	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id string) error {
	query := `
		UPDATE refresh_token
		SET revoked_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_token
		SET revoked_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}
