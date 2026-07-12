package repository

import (
	"context"
	"time"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type OrderFilter struct {
	Page          int
	PerPage       int
	Status        string
	From          *time.Time
	To            *time.Time
	CustomerEmail string
}

type DailyRevenuePoint struct {
	Date    string `db:"date"`
	Revenue int64  `db:"revenue"`
}

type TopProductPoint struct {
	ProductName  string `db:"product_name"`
	QuantitySold int64  `db:"quantity_sold"`
	Revenue      int64  `db:"revenue"`
}

type SalesSummary struct {
	TotalRevenue int64
	TotalOrders  int64
	StatusCounts map[string]int64
	DailyRevenue []DailyRevenuePoint
	TopProducts  []TopProductPoint
}

type CustomerStats struct {
	OrderCount int64
	TotalSpent int64
}

type OrderRepository interface {
	FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	// FindByIdempotencyKey looks up an order previously created with the same
	// client-supplied key, returning ErrOrderNotFound when none exists.
	FindByIdempotencyKey(ctx context.Context, key string) (*entity.Order, error)
	Create(ctx context.Context, order *entity.Order) error
	UpdateInvoiceURL(ctx context.Context, id string, invoiceURL string) error
	UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error
	UpdatePaymentStatus(ctx context.Context, id string, status entity.PaymentStatus) error
	// MarkPaidIfUnpaid atomically transitions an order to "paid" only if it
	// isn't already paid, returning whether this call was the one that made
	// the change. Callers use this to fire paid-side-effects (notifications)
	// exactly once even when the same payment confirmation arrives more than
	// once concurrently (webhook retries, or webhook + manual status check
	// racing each other).
	MarkPaidIfUnpaid(ctx context.Context, id string) (bool, error)
	// FindExpiredUnpaidOrders returns still-unpaid orders whose payment
	// deadline (payment_expiry) has passed as of `now`, items included so the
	// caller can release each item's stock reservation.
	FindExpiredUnpaidOrders(ctx context.Context, now time.Time) ([]entity.Order, error)
	// ExpireIfUnpaid atomically transitions an order to "expired" only if it
	// is still "unpaid" — the same guard pattern as MarkPaidIfUnpaid, so a
	// payment that lands (webhook or manual check) at the exact moment the
	// expiry sweep runs can never be overwritten back to expired.
	ExpireIfUnpaid(ctx context.Context, id string) (bool, error)
	GetSalesSummary(ctx context.Context, from, to time.Time) (*SalesSummary, error)
	GetCustomerStats(ctx context.Context, email string) (orderCount int64, totalSpent int64, err error)
	GetCustomerStatsBulk(ctx context.Context, emails []string) (map[string]CustomerStats, error)
}
