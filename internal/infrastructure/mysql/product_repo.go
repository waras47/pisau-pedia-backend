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
	if filter.Badge != "" {
		conditions = append(conditions, "p.badge = ?")
		args = append(args, filter.Badge)
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
	case "rating":
		return "p.rating_avg DESC, p.review_count DESC"
	case "bestseller":
		return "COALESCE((SELECT SUM(oi.quantity) FROM order_items oi INNER JOIN orders o ON o.id = oi.order_id WHERE oi.product_id = p.id AND o.status NOT IN ('cancelled','expired')), 0) DESC, p.rating_avg DESC"
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

	if err := r.attachPrimaryImages(ctx, products); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// attachPrimaryImages fetches the first image per product in a single
// batched query (window function, not a join) and sets Product.Image —
// enough for list-view thumbnails without pulling the full Images payload.
func (r *productRepository) attachPrimaryImages(ctx context.Context, products []entity.Product) error {
	if len(products) == 0 {
		return nil
	}

	ids := make([]string, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}

	query, args, err := sqlx.In(`
		SELECT product_id, url FROM (
			SELECT product_id, url,
				ROW_NUMBER() OVER (PARTITION BY product_id ORDER BY sort_order) AS rn
			FROM product_images
			WHERE product_id IN (?)
		) ranked
		WHERE rn = 1
	`, ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)

	var rows []struct {
		ProductID string `db:"product_id"`
		URL       string `db:"url"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return err
	}

	imageByProduct := make(map[string]string, len(rows))
	for _, row := range rows {
		imageByProduct[row.ProductID] = row.URL
	}
	for i := range products {
		if url, ok := imageByProduct[products[i].ID]; ok {
			products[i].Image = &url
		}
	}
	return nil
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

func (r *productRepository) ExistsBySKU(ctx context.Context, sku string, excludeID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM products WHERE sku = ? AND id != ?`, sku, excludeID)
	return count > 0, err
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO products (id, category_id, name, slug, sku, description, description_en, care_instructions, price, compare_at_price, currency, maker, badge, stock, weight, is_active)
		VALUES (:id, :category_id, :name, :slug, :sku, :description, :description_en, :care_instructions, :price, :compare_at_price, :currency, :maker, :badge, :stock, :weight, :is_active)
	`, product)
	if err != nil {
		return err
	}

	for i := range product.Images {
		product.Images[i].ProductID = product.ID
		product.Images[i].SortOrder = uint(i)
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO product_images (id, product_id, url, alt_text, angle, sort_order)
			VALUES (:id, :product_id, :url, :alt_text, :angle, :sort_order)
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
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		UPDATE products SET
			category_id = :category_id,
			name = :name,
			sku = :sku,
			description = :description,
			description_en = :description_en,
			care_instructions = :care_instructions,
			price = :price,
			compare_at_price = :compare_at_price,
			maker = :maker,
			badge = :badge,
			stock = :stock,
			weight = :weight,
			is_active = :is_active
		WHERE id = :id
	`, product)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE product_id = ?`, product.ID); err != nil {
		return err
	}
	for i := range product.Images {
		product.Images[i].ProductID = product.ID
		product.Images[i].SortOrder = uint(i)
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO product_images (id, product_id, url, alt_text, angle, sort_order)
			VALUES (:id, :product_id, :url, :alt_text, :angle, :sort_order)
		`, product.Images[i]); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_specs WHERE product_id = ?`, product.ID); err != nil {
		return err
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_highlights WHERE product_id = ?`, product.ID); err != nil {
		return err
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

func (r *productRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = ?`, id)
	return err
}

func (r *productRepository) DecrementStockIfAvailable(ctx context.Context, productID string, qty uint) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`,
		qty, productID, qty,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *productRepository) RestoreStock(ctx context.Context, productID string, qty uint) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET stock = stock + ? WHERE id = ?`, qty, productID)
	return err
}

func (r *productRepository) GetInventorySummary(ctx context.Context, lowStockThreshold uint) (*repository.InventorySummary, error) {
	summary := &repository.InventorySummary{}

	if err := r.db.GetContext(ctx, &summary.TotalProducts, `SELECT COUNT(*) FROM products WHERE is_active = 1`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &summary.TotalStockValue, `
		SELECT COALESCE(SUM(price * stock), 0) FROM products WHERE is_active = 1
	`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &summary.LowStockCount, `
		SELECT COUNT(*) FROM products WHERE is_active = 1 AND stock > 0 AND stock <= ?
	`, lowStockThreshold); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &summary.OutOfStockCount, `
		SELECT COUNT(*) FROM products WHERE is_active = 1 AND stock = 0
	`); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &summary.CategoryBreakdown, `
		SELECT COALESCE(c.name, 'Tanpa Kategori') AS category_name,
			COUNT(*) AS product_count,
			COALESCE(SUM(p.stock), 0) AS total_stock
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.is_active = 1
		GROUP BY category_name
		ORDER BY category_name
	`); err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *productRepository) UpdateRatingStats(ctx context.Context, productID string, ratingAvg float64, reviewCount uint) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE products SET rating_avg = ?, review_count = ? WHERE id = ?
	`, ratingAvg, reviewCount, productID)
	return err
}
