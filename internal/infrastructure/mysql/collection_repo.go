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

type collectionRepository struct {
	db *sqlx.DB
}

func NewCollectionRepository(db *sqlx.DB) repository.CollectionRepository {
	return &collectionRepository{db: db}
}

func (r *collectionRepository) attachCategories(ctx context.Context, collections []entity.Collection) error {
	if len(collections) == 0 {
		return nil
	}

	ids := make([]string, len(collections))
	idxMap := make(map[string]int, len(collections))
	for i := range collections {
		ids[i] = collections[i].ID
		idxMap[collections[i].ID] = i
		collections[i].Categories = []entity.Category{}
	}

	query, args, err := sqlx.In(`
		SELECT c.*, cc.collection_id
		FROM categories c
		INNER JOIN collection_categories cc ON cc.category_id = c.id
		WHERE cc.collection_id IN (?)
		ORDER BY c.name
	`, ids)
	if err != nil {
		return err
	}

	type row struct {
		entity.Category
		CollectionID string `db:"collection_id"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(query), args...); err != nil {
		return err
	}

	for _, row := range rows {
		if idx, ok := idxMap[row.CollectionID]; ok {
			collections[idx].Categories = append(collections[idx].Categories, row.Category)
		}
	}
	return nil
}

func (r *collectionRepository) FindAll(ctx context.Context) ([]entity.Collection, error) {
	var collections []entity.Collection
	if err := r.db.SelectContext(ctx, &collections, `SELECT * FROM collections ORDER BY sort_order, name`); err != nil {
		return nil, err
	}
	if err := r.attachCategories(ctx, collections); err != nil {
		return nil, err
	}
	return collections, nil
}

func (r *collectionRepository) FindBySlug(ctx context.Context, slug string) (*entity.Collection, error) {
	var collection entity.Collection
	err := r.db.GetContext(ctx, &collection, `SELECT * FROM collections WHERE slug = ?`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCollectionNotFound
	}
	if err != nil {
		return nil, err
	}
	cols := []entity.Collection{collection}
	if err := r.attachCategories(ctx, cols); err != nil {
		return nil, err
	}
	return &cols[0], nil
}

func (r *collectionRepository) FindByID(ctx context.Context, id string) (*entity.Collection, error) {
	var collection entity.Collection
	err := r.db.GetContext(ctx, &collection, `SELECT * FROM collections WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCollectionNotFound
	}
	if err != nil {
		return nil, err
	}
	cols := []entity.Collection{collection}
	if err := r.attachCategories(ctx, cols); err != nil {
		return nil, err
	}
	return &cols[0], nil
}

func (r *collectionRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM collections WHERE slug = ?`, slug)
	return count > 0, err
}

func (r *collectionRepository) Create(ctx context.Context, collection *entity.Collection) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO collections (id, name, slug, description, image_url, sort_order)
		VALUES (:id, :name, :slug, :description, :image_url, :sort_order)
	`, collection)
	return err
}

func (r *collectionRepository) Update(ctx context.Context, collection *entity.Collection) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE collections SET
			name = :name,
			description = :description,
			image_url = :image_url,
			sort_order = :sort_order
		WHERE id = :id
	`, collection)
	return err
}

func (r *collectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	return err
}

func (r *collectionRepository) SetCategories(ctx context.Context, collectionID string, categoryIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM collection_categories WHERE collection_id = ?`, collectionID); err != nil {
		return err
	}

	if len(categoryIDs) > 0 {
		valueStrings := make([]string, len(categoryIDs))
		valueArgs := make([]interface{}, 0, len(categoryIDs)*2)
		for i, catID := range categoryIDs {
			valueStrings[i] = "(?, ?)"
			valueArgs = append(valueArgs, collectionID, catID)
		}
		query := fmt.Sprintf(`INSERT INTO collection_categories (collection_id, category_id) VALUES %s`, strings.Join(valueStrings, ", "))
		if _, err := tx.ExecContext(ctx, query, valueArgs...); err != nil {
			return err
		}
	}

	return tx.Commit()
}
