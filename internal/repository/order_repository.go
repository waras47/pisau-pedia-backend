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
	// UserID restricts results to orders placed while signed in as this
	// user — used for the customer-facing "Pesanan Saya" list. Guest orders
	// (user_id NULL) never match, by design.
	UserID string
	// Search matches against order id, customer name, or customer email —
	// used by the admin global search, not the customer-facing order list.
	Search string
}

type DailyRevenuePoint struct {
	// Date scans as time.Time, not string, because DB_PARAMS has
	// parseTime=true — MySQL DATE columns come back as time.Time under
	// that driver setting. Format it explicitly at the DTO boundary
	// (see ToSalesReportResponse) rather than typing this string and
	// letting database/sql silently reformat it to a full RFC3339
	// timestamp, which broke the frontend's plain-date lookup.
	Date    time.Time `db:"date"`
	Revenue int64     `db:"revenue"`
}

type TopProductPoint struct {
	ProductName  string `db:"product_name"`
	QuantitySold int64  `db:"quantity_sold"`
	Revenue      int64  `db:"revenue"`
}

type SalesSummary struct {
	TotalRevenue int64
	TotalOrders  int64
	// PaidOrders is the subset of TotalOrders with payment_status = 'paid'
	// — the count that actually backs TotalRevenue/DailyRevenue/TopProducts,
	// which are all paid-only. Surfaced separately so the dashboard can show
	// order volume and paid order count without them looking inconsistent.
	PaidOrders   int64
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
	// UpdatePaymentProof stores the URL of a customer-uploaded payment proof
	// (receipt/screenshot) — used by the manual/static payment flow, where
	// there's no gateway webhook to confirm payment automatically.
	UpdatePaymentProof(ctx context.Context, id string, url string) error
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
	// ConfirmReceivedIfEligible sets customer_confirmed_at for an order the
	// given user owns, atomically and only once — a second call (double
	// click, two tabs) is a harmless no-op rather than a duplicate
	// notification. Returns false (not an error) if the order doesn't
	// belong to this user, isn't paid, or is still pending/cancelled.
	ConfirmReceivedIfEligible(ctx context.Context, orderID, userID string) (bool, error)
	GetSalesSummary(ctx context.Context, from, to time.Time) (*SalesSummary, error)
	GetCustomerStats(ctx context.Context, email string) (orderCount int64, totalSpent int64, err error)
	GetCustomerStatsBulk(ctx context.Context, emails []string) (map[string]CustomerStats, error)
}
