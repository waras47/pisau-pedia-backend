package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type emailVerificationRepository struct {
	db *sqlx.DB
}

func NewEmailVerificationRepository(db *sqlx.DB) repository.EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, token *entity.EmailVerificationToken) error {
	query := `INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *emailVerificationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.EmailVerificationToken, error) {
	var token entity.EmailVerificationToken
	query := `SELECT id, user_id, token_hash, expires_at, created_at FROM email_verification_tokens WHERE token_hash = ?`
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrVerificationTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = ?`, userID)
	return err
}
