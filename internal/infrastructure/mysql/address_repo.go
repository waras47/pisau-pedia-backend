package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type addressRepository struct {
	db *sqlx.DB
}

func NewAddressRepository(db *sqlx.DB) repository.AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *entity.Address) error {
	query := `
		INSERT INTO addresses (id, user_id, label, full_name, phone, address_line, city, province, postal_code, is_default)
		VALUES (:id, :user_id, :label, :full_name, :phone, :address_line, :city, :province, :postal_code, :is_default)
	`
	_, err := r.db.NamedExecContext(ctx, query, address)
	return err
}

func (r *addressRepository) FindByUserID(ctx context.Context, userID string) ([]entity.Address, error) {
	var addresses []entity.Address
	err := r.db.SelectContext(ctx, &addresses, `SELECT * FROM addresses WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *addressRepository) FindByID(ctx context.Context, id string) (*entity.Address, error) {
	var address entity.Address
	err := r.db.GetContext(ctx, &address, `SELECT * FROM addresses WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrAddressNotFound
	}
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *addressRepository) Update(ctx context.Context, address *entity.Address) error {
	query := `
		UPDATE addresses SET
			label = :label,
			full_name = :full_name,
			phone = :phone,
			address_line = :address_line,
			city = :city,
			province = :province,
			postal_code = :postal_code,
			is_default = :is_default
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, address)
	return err
}

func (r *addressRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = ?`, id)
	return err
}

func (r *addressRepository) UnsetDefaultForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE addresses SET is_default = 0 WHERE user_id = ? AND is_default = 1`, userID)
	return err
}
