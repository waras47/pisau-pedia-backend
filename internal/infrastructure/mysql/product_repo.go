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

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) repository.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) buildListFilter(filter repository.ProductFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "p.is_active = 1")

	if filter.CategorySlug != "" {
		conditions = append(conditions, "c.slug = ?")
		args = append(args, filter.CategorySlug)
	}
	if filter.Search != "" {
		conditions = append(conditions, "p.name LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	return strings.Join(conditions, " AND "), args
}

func sortClause(sort string) string {
	switch sort {
	case "price_asc":
		return "p.price ASC"
	case "price_desc":
		return "p.price DESC"
	case "newest":
		return "p.created_at DESC"
	default:
		return "p.created_at DESC"
	}
}

func (r *productRepository) FindAll(ctx context.Context, filter repository.ProductFilter) ([]entity.Product, int64, error) {
	where, args := r.buildListFilter(filter)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s
	`, where)

	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT p.*, c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, where, sortClause(filter.Sort))

	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var products []entity.Product
	if err := r.db.SelectContext(ctx, &products, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	var product entity.Product
	query := `
		SELECT p.*, c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug = ?
	`
	if err := r.db.GetContext(ctx, &product, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}

	if err := r.loadChildren(ctx, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	var product entity.Product
	query := `
		SELECT p.*, c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = ?
	`
	if err := r.db.GetContext(ctx, &product, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}

	if err := r.loadChildren(ctx, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) loadChildren(ctx context.Context, product *entity.Product) error {
	if err := r.db.SelectContext(ctx, &product.Images,
		`SELECT * FROM product_images WHERE product_id = ? ORDER BY sort_order`, product.ID); err != nil {
		return err
	}
	if err := r.db.SelectContext(ctx, &product.Specs,
		`SELECT * FROM product_specs WHERE product_id = ? ORDER BY sort_order`, product.ID); err != nil {
		return err
	}
	if err := r.db.SelectContext(ctx, &product.Highlights,
		`SELECT * FROM product_highlights WHERE product_id = ? ORDER BY sort_order`, product.ID); err != nil {
		return err
	}
	return nil
}

func (r *productRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM products WHERE slug = ?`, slug)
	return count > 0, err
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO products (id, category_id, name, slug, description, price, compare_at_price, currency, maker, badge, stock, is_active)
		VALUES (:id, :category_id, :name, :slug, :description, :price, :compare_at_price, :currency, :maker, :badge, :stock, :is_active)
	`, product)
	if err != nil {
		return err
	}

	for i := range product.Images {
		product.Images[i].ProductID = product.ID
		product.Images[i].SortOrder = uint(i)
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO product_images (id, product_id, url, alt_text, sort_order)
			VALUES (:id, :product_id, :url, :alt_text, :sort_order)
		`, product.Images[i]); err != nil {
			return err
		}
	}

	for i := range product.Specs {
		product.Specs[i].ProductID = product.ID
		product.Specs[i].SortOrder = uint(i)
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO product_specs (id, product_id, label, value, sort_order)
			VALUES (:id, :product_id, :label, :value, :sort_order)
		`, product.Specs[i]); err != nil {
			return err
		}
	}

	for i := range product.Highlights {
		product.Highlights[i].ProductID = product.ID
		product.Highlights[i].SortOrder = uint(i)
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO product_highlights (id, product_id, highlight, sort_order)
			VALUES (:id, :product_id, :highlight, :sort_order)
		`, product.Highlights[i]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	query := `
		UPDATE products SET
			category_id = :category_id,
			name = :name,
			description = :description,
			price = :price,
			compare_at_price = :compare_at_price,
			maker = :maker,
			badge = :badge,
			stock = :stock,
			is_active = :is_active
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = ?`, id)
	return err
}
