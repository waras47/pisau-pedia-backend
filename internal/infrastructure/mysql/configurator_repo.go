package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

// --- Shape ---

type configuratorShapeRepo struct{ db *sqlx.DB }

func NewConfiguratorShapeRepository(db *sqlx.DB) repository.ConfiguratorShapeRepository {
	return &configuratorShapeRepo{db: db}
}

func (r *configuratorShapeRepo) FindAll(ctx context.Context) ([]entity.ConfiguratorShape, error) {
	var out []entity.ConfiguratorShape
	return out, r.db.SelectContext(ctx, &out, `SELECT * FROM configurator_shapes ORDER BY sort_order, name`)
}

func (r *configuratorShapeRepo) FindByID(ctx context.Context, id string) (*entity.ConfiguratorShape, error) {
	var s entity.ConfiguratorShape
	err := r.db.GetContext(ctx, &s, `SELECT * FROM configurator_shapes WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrConfiguratorShapeNotFound
	}
	return &s, err
}

func (r *configuratorShapeRepo) Create(ctx context.Context, s *entity.ConfiguratorShape) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO configurator_shapes (id, name, category, description, image_url, sort_order)
		VALUES (:id, :name, :category, :description, :image_url, :sort_order)`, s)
	return err
}

func (r *configuratorShapeRepo) Update(ctx context.Context, s *entity.ConfiguratorShape) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE configurator_shapes SET name=:name, category=:category, description=:description,
		image_url=:image_url, sort_order=:sort_order WHERE id=:id`, s)
	return err
}

func (r *configuratorShapeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM configurator_shapes WHERE id = ?`, id)
	return err
}

// --- Blade ---

type configuratorBladeRepo struct{ db *sqlx.DB }

func NewConfiguratorBladeRepository(db *sqlx.DB) repository.ConfiguratorBladeRepository {
	return &configuratorBladeRepo{db: db}
}

func (r *configuratorBladeRepo) FindAll(ctx context.Context) ([]entity.ConfiguratorBlade, error) {
	var out []entity.ConfiguratorBlade
	return out, r.db.SelectContext(ctx, &out, `SELECT * FROM configurator_blades ORDER BY sort_order, name`)
}

func (r *configuratorBladeRepo) FindByShapeID(ctx context.Context, shapeID string) ([]entity.ConfiguratorBlade, error) {
	var out []entity.ConfiguratorBlade
	return out, r.db.SelectContext(ctx, &out, `SELECT * FROM configurator_blades WHERE shape_id = ? ORDER BY sort_order, name`, shapeID)
}

func (r *configuratorBladeRepo) FindByID(ctx context.Context, id string) (*entity.ConfiguratorBlade, error) {
	var b entity.ConfiguratorBlade
	err := r.db.GetContext(ctx, &b, `SELECT * FROM configurator_blades WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrConfiguratorBladeNotFound
	}
	return &b, err
}

func (r *configuratorBladeRepo) Create(ctx context.Context, b *entity.ConfiguratorBlade) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO configurator_blades (id, shape_id, name, steel, length_mm, price, compare_at_price, description, specifications, image_url, sort_order)
		VALUES (:id, :shape_id, :name, :steel, :length_mm, :price, :compare_at_price, :description, :specifications, :image_url, :sort_order)`, b)
	return err
}

func (r *configuratorBladeRepo) Update(ctx context.Context, b *entity.ConfiguratorBlade) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE configurator_blades SET shape_id=:shape_id, name=:name, steel=:steel, length_mm=:length_mm,
		price=:price, compare_at_price=:compare_at_price, description=:description, specifications=:specifications,
		image_url=:image_url, sort_order=:sort_order WHERE id=:id`, b)
	return err
}

func (r *configuratorBladeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM configurator_blades WHERE id = ?`, id)
	return err
}

// --- Handle ---

type configuratorHandleRepo struct{ db *sqlx.DB }

func NewConfiguratorHandleRepository(db *sqlx.DB) repository.ConfiguratorHandleRepository {
	return &configuratorHandleRepo{db: db}
}

func (r *configuratorHandleRepo) FindAll(ctx context.Context) ([]entity.ConfiguratorHandle, error) {
	var out []entity.ConfiguratorHandle
	return out, r.db.SelectContext(ctx, &out, `SELECT * FROM configurator_handles ORDER BY sort_order, name`)
}

func (r *configuratorHandleRepo) FindByID(ctx context.Context, id string) (*entity.ConfiguratorHandle, error) {
	var h entity.ConfiguratorHandle
	err := r.db.GetContext(ctx, &h, `SELECT * FROM configurator_handles WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrConfiguratorHandleNotFound
	}
	return &h, err
}

func (r *configuratorHandleRepo) Create(ctx context.Context, h *entity.ConfiguratorHandle) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO configurator_handles (id, name, material, price_delta, image_url, sort_order)
		VALUES (:id, :name, :material, :price_delta, :image_url, :sort_order)`, h)
	return err
}

func (r *configuratorHandleRepo) Update(ctx context.Context, h *entity.ConfiguratorHandle) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE configurator_handles SET name=:name, material=:material, price_delta=:price_delta,
		image_url=:image_url, sort_order=:sort_order WHERE id=:id`, h)
	return err
}

func (r *configuratorHandleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM configurator_handles WHERE id = ?`, id)
	return err
}

// --- Accessory ---

type configuratorAccessoryRepo struct{ db *sqlx.DB }

func NewConfiguratorAccessoryRepository(db *sqlx.DB) repository.ConfiguratorAccessoryRepository {
	return &configuratorAccessoryRepo{db: db}
}

func (r *configuratorAccessoryRepo) FindAll(ctx context.Context) ([]entity.ConfiguratorAccessory, error) {
	var out []entity.ConfiguratorAccessory
	return out, r.db.SelectContext(ctx, &out, `SELECT * FROM configurator_accessories ORDER BY sort_order, name`)
}

func (r *configuratorAccessoryRepo) FindByID(ctx context.Context, id string) (*entity.ConfiguratorAccessory, error) {
	var a entity.ConfiguratorAccessory
	err := r.db.GetContext(ctx, &a, `SELECT * FROM configurator_accessories WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrConfiguratorAccessoryNotFound
	}
	return &a, err
}

func (r *configuratorAccessoryRepo) Create(ctx context.Context, a *entity.ConfiguratorAccessory) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO configurator_accessories (id, name, price, image_url, sort_order)
		VALUES (:id, :name, :price, :image_url, :sort_order)`, a)
	return err
}

func (r *configuratorAccessoryRepo) Update(ctx context.Context, a *entity.ConfiguratorAccessory) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE configurator_accessories SET name=:name, price=:price,
		image_url=:image_url, sort_order=:sort_order WHERE id=:id`, a)
	return err
}

func (r *configuratorAccessoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM configurator_accessories WHERE id = ?`, id)
	return err
}
