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
	Create(ctx context.Context, order *entity.Order) error
	UpdateInvoiceURL(ctx context.Context, id string, invoiceURL string) error
	UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error
	UpdatePaymentStatus(ctx context.Context, id string, status entity.PaymentStatus) error
	GetSalesSummary(ctx context.Context, from, to time.Time) (*SalesSummary, error)
	GetCustomerStats(ctx context.Context, email string) (orderCount int64, totalSpent int64, err error)
	GetCustomerStatsBulk(ctx context.Context, emails []string) (map[string]CustomerStats, error)
}
