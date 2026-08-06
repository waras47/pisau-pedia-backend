package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, token *entity.EmailVerificationToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.EmailVerificationToken, error)
	DeleteByUserID(ctx context.Context, userID string) error
}
