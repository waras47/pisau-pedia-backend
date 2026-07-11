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

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

const userColumns = `
	id, email, password_hash, full_name, phone, role,
	avatar_url, is_active, email_verified_at, created_at, updated_at
`

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
	query := `SELECT ` + userColumns + ` FROM users WHERE email = ?`
	err := r.db.GetContext(ctx, &user, query, email)
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
	query := `SELECT ` + userColumns + ` FROM users WHERE id = ?`
	err := r.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, filter repository.UserFilter) ([]entity.User, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Role != "" {
		conditions = append(conditions, "role = ?")
		args = append(args, filter.Role)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(full_name LIKE ? OR email LIKE ?)")
		q := "%" + filter.Search + "%"
		args = append(args, q, q)
	}

	where := "1=1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT %s FROM users
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userColumns, where)
	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var users []entity.User
	if err := r.db.SelectContext(ctx, &users, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users SET
			full_name = :full_name,
			phone = :phone,
			role = :role,
			avatar_url = :avatar_url,
			password_hash = :password_hash,
			is_active = :is_active
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}
