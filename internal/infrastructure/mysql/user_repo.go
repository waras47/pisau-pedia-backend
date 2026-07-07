package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, phone, role, avatar_url, is_active)
		VALUES (:id, :email, :password_hash, :full_name, :phone, :role, :avatar_url, :is_active)
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = ?`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users SET
			full_name = :full_name,
			phone = :phone,
			avatar_url = :avatar_url,
			password_hash = :password_hash,
			is_active = :is_active
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}
