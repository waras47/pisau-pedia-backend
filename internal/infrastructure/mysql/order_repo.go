package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) FindAll(ctx context.Context, filter repository.OrderFilter) ([]entity.Order, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.From != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.From)
	}
	if filter.To != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.To)
	}
	if filter.CustomerEmail != "" {
		conditions = append(conditions, "customer_email = ?")
		args = append(args, filter.CustomerEmail)
	}

	where := "1=1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM orders WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT * FROM orders
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var orders []entity.Order
	if err := r.db.SelectContext(ctx, &orders, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.GetContext(ctx, &order, `SELECT * FROM orders WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &order.Items,
		`SELECT * FROM order_items WHERE order_id = ?`, order.ID); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.GetContext(ctx, &order, `SELECT * FROM orders WHERE idempotency_key = ?`, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &order.Items,
		`SELECT * FROM order_items WHERE order_id = ?`, order.ID); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO orders (
			id, idempotency_key, user_id, status, payment_status,
			payment_provider, payment_id, payment_type, payment_channel,
			payment_va_number, payment_qr_string, payment_url, payment_expiry,
			customer_name, customer_email, customer_phone,
			shipping_address, shipping_city, shipping_province, shipping_postal_code,
			subtotal, total, currency, coupon_code, discount_amount,
			shipping_cost, shipping_courier, shipping_service, shipping_etd, destination_id,
			xendit_external_id, xendit_invoice_url
		) VALUES (
			:id, :idempotency_key, :user_id, :status, :payment_status,
			:payment_provider, :payment_id, :payment_type, :payment_channel,
			:payment_va_number, :payment_qr_string, :payment_url, :payment_expiry,
			:customer_name, :customer_email, :customer_phone,
			:shipping_address, :shipping_city, :shipping_province, :shipping_postal_code,
			:subtotal, :total, :currency, :coupon_code, :discount_amount,
			:shipping_cost, :shipping_courier, :shipping_service, :shipping_etd, :destination_id,
			:xendit_external_id, :xendit_invoice_url
		)
	`, order)
	if err != nil {
		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			// Two concurrent double-submits with the same idempotency key
			// both passed the pre-insert FindByIdempotencyKey check before
			// either committed — the UNIQUE index is the real guard here.
			return repository.ErrDuplicateEntry
		}
		return err
	}

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO order_items (id, order_id, product_id, product_name, product_slug, price, quantity, subtotal)
			VALUES (:id, :order_id, :product_id, :product_name, :product_slug, :price, :quantity, :subtotal)
		`, order.Items[i]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepository) UpdateInvoiceURL(ctx context.Context, id string, invoiceURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE orders SET xendit_invoice_url = ? WHERE id = ?`, invoiceURL, id)
	return err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE orders SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id string, status entity.PaymentStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE orders SET payment_status = ? WHERE id = ?`, status, id)
	return err
}

func (r *orderRepository) MarkPaidIfUnpaid(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET payment_status = ? WHERE id = ? AND payment_status != ?`,
		entity.PaymentStatusPaid, id, entity.PaymentStatusPaid,
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

// FindExpiredUnpaidOrders filters in Go rather than SQL because
// payment_expiry is stored as the RFC3339 string Komerce returns, not a
// proper DATETIME column — lexicographic comparison isn't reliably safe
// across timezone-offset formatting differences.
func (r *orderRepository) FindExpiredUnpaidOrders(ctx context.Context, now time.Time) ([]entity.Order, error) {
	var candidates []entity.Order
	if err := r.db.SelectContext(ctx, &candidates, `
		SELECT * FROM orders
		WHERE payment_status = ? AND payment_expiry IS NOT NULL AND payment_expiry != ''
	`, entity.PaymentStatusUnpaid); err != nil {
		return nil, err
	}

	var expired []entity.Order
	for _, o := range candidates {
		deadline, err := time.Parse(time.RFC3339, *o.PaymentExpiry)
		if err != nil || !now.After(deadline) {
			continue
		}
		if err := r.db.SelectContext(ctx, &o.Items, `SELECT * FROM order_items WHERE order_id = ?`, o.ID); err != nil {
			return nil, err
		}
		expired = append(expired, o)
	}
	return expired, nil
}

func (r *orderRepository) ExpireIfUnpaid(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET payment_status = ? WHERE id = ? AND payment_status = ?`,
		entity.PaymentStatusExpired, id, entity.PaymentStatusUnpaid,
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

func (r *orderRepository) GetSalesSummary(ctx context.Context, from, to time.Time) (*repository.SalesSummary, error) {
	summary := &repository.SalesSummary{StatusCounts: map[string]int64{}}

	if err := r.db.GetContext(ctx, &summary.TotalRevenue, `
		SELECT COALESCE(SUM(total), 0) FROM orders
		WHERE payment_status = 'paid' AND created_at BETWEEN ? AND ?
	`, from, to); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &summary.TotalOrders, `
		SELECT COUNT(*) FROM orders WHERE created_at BETWEEN ? AND ?
	`, from, to); err != nil {
		return nil, err
	}

	type statusRow struct {
		Status string `db:"status"`
		Count  int64  `db:"count"`
	}
	var statusRows []statusRow
	if err := r.db.SelectContext(ctx, &statusRows, `
		SELECT status, COUNT(*) AS count FROM orders
		WHERE created_at BETWEEN ? AND ?
		GROUP BY status
	`, from, to); err != nil {
		return nil, err
	}
	for _, s := range statusRows {
		summary.StatusCounts[s.Status] = s.Count
	}

	if err := r.db.SelectContext(ctx, &summary.DailyRevenue, `
		SELECT DATE(created_at) AS date, COALESCE(SUM(total), 0) AS revenue
		FROM orders
		WHERE payment_status = 'paid' AND created_at BETWEEN ? AND ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, from, to); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &summary.TopProducts, `
		SELECT oi.product_name AS product_name, SUM(oi.quantity) AS quantity_sold, SUM(oi.subtotal) AS revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.payment_status = 'paid' AND o.created_at BETWEEN ? AND ?
		GROUP BY oi.product_name
		ORDER BY quantity_sold DESC
		LIMIT 5
	`, from, to); err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *orderRepository) GetCustomerStats(ctx context.Context, email string) (int64, int64, error) {
	var stats struct {
		OrderCount int64 `db:"order_count"`
		TotalSpent int64 `db:"total_spent"`
	}
	err := r.db.GetContext(ctx, &stats, `
		SELECT COUNT(*) AS order_count, COALESCE(SUM(total), 0) AS total_spent
		FROM orders
		WHERE customer_email = ? AND payment_status = 'paid'
	`, email)
	if err != nil {
		return 0, 0, err
	}
	return stats.OrderCount, stats.TotalSpent, nil
}

func (r *orderRepository) GetCustomerStatsBulk(ctx context.Context, emails []string) (map[string]repository.CustomerStats, error) {
	result := make(map[string]repository.CustomerStats, len(emails))
	if len(emails) == 0 {
		return result, nil
	}

	query, args, err := sqlx.In(`
		SELECT customer_email, COUNT(*) AS order_count, COALESCE(SUM(total), 0) AS total_spent
		FROM orders
		WHERE customer_email IN (?) AND payment_status = 'paid'
		GROUP BY customer_email
	`, emails)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var rows []struct {
		CustomerEmail string `db:"customer_email"`
		OrderCount    int64  `db:"order_count"`
		TotalSpent    int64  `db:"total_spent"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.CustomerEmail] = repository.CustomerStats{OrderCount: row.OrderCount, TotalSpent: row.TotalSpent}
	}
	return result, nil
}
