package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/rajaongkir"
)

var ErrEmptyOrder = errors.New("order must have at least one valid item")

type OrderItemInput struct {
	ProductSlug string
	Quantity    uint
}

type CreateOrderInput struct {
	CustomerName       string
	CustomerEmail      string
	CustomerPhone      *string
	ShippingAddress    string
	ShippingCity       string
	ShippingProvince   *string
	ShippingPostalCode string
	CouponCode         string
	Items              []OrderItemInput
}

type OrderListInput struct {
	Page          int
	PerPage       int
	Status        string
	CustomerEmail string
}

type OrderListResult struct {
	Orders     []entity.Order
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

type OrderUsecase struct {
	orderRepo           repository.OrderRepository
	productRepo         repository.ProductRepository
	paymentGateway      PaymentGateway
	couponUsecase       *CouponUsecase
	notificationUsecase *NotificationUsecase
	shippingClient      *rajaongkir.Client
}

func NewOrderUsecase(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, paymentGateway PaymentGateway, couponUsecase *CouponUsecase, notificationUsecase *NotificationUsecase, shippingClient *rajaongkir.Client) *OrderUsecase {
	return &OrderUsecase{orderRepo: orderRepo, productRepo: productRepo, paymentGateway: paymentGateway, couponUsecase: couponUsecase, notificationUsecase: notificationUsecase, shippingClient: shippingClient}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*entity.Order, error) {
	var items []entity.OrderItem
	var subtotal int64

	// Prices are recomputed from the product catalog, never trusted from the
	// client — the same defensive pattern the old Next.js checkout route used.
	for _, i := range input.Items {
		if i.Quantity == 0 {
			continue
		}
		product, err := u.productRepo.FindBySlug(ctx, i.ProductSlug)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				continue
			}
			return nil, err
		}
		lineSubtotal := product.Price * int64(i.Quantity)
		subtotal += lineSubtotal
		items = append(items, entity.OrderItem{
			ID:          uuid.New().String(),
			ProductID:   &product.ID,
			ProductName: product.Name,
			ProductSlug: product.Slug,
			Price:       product.Price,
			Quantity:    i.Quantity,
			Subtotal:    lineSubtotal,
		})
	}

	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	var couponCode *string
	var couponID string
	var discountAmount int64
	if input.CouponCode != "" {
		result, err := u.couponUsecase.ValidateCoupon(ctx, input.CouponCode, subtotal)
		if err != nil {
			return nil, err
		}
		couponCode = &input.CouponCode
		couponID = result.Coupon.ID
		discountAmount = result.DiscountAmount
	}

	orderID := uuid.New().String()
	now := time.Now()
	order := &entity.Order{
		ID:                 orderID,
		Status:             entity.OrderStatusPending,
		PaymentStatus:      entity.PaymentStatusUnpaid,
		CreatedAt:          now,
		UpdatedAt:          now,
		CustomerName:       input.CustomerName,
		CustomerEmail:      input.CustomerEmail,
		CustomerPhone:      input.CustomerPhone,
		ShippingAddress:    input.ShippingAddress,
		ShippingCity:       input.ShippingCity,
		ShippingProvince:   input.ShippingProvince,
		ShippingPostalCode: input.ShippingPostalCode,
		Subtotal:           subtotal,
		Total:              subtotal - discountAmount,
		Currency:           "IDR",
		CouponCode:         couponCode,
		DiscountAmount:     discountAmount,
		XenditExternalID:   &orderID,
		Items:              items,
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	if couponID != "" {
		if err := u.couponUsecase.couponRepo.IncrementUsage(ctx, couponID); err != nil {
			return nil, err
		}
	}

	invoice, err := u.paymentGateway.CreateInvoice(ctx, CreateInvoiceInput{
		ExternalID:    orderID,
		Amount:        order.Total,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		Description:   fmt.Sprintf("Pisau Pedia order — %d item(s)", len(items)),
	})
	if err != nil {
		return nil, err
	}

	if err := u.orderRepo.UpdateInvoiceURL(ctx, orderID, invoice.InvoiceURL); err != nil {
		return nil, err
	}
	order.XenditInvoiceURL = &invoice.InvoiceURL

	// Notification failures shouldn't block order creation.
	_ = u.notificationUsecase.NotifyOrderCreated(ctx, order)

	return order, nil
}

func (u *OrderUsecase) ListOrders(ctx context.Context, input OrderListInput) (*OrderListResult, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	orders, total, err := u.orderRepo.FindAll(ctx, repository.OrderFilter{
		Page:          page,
		PerPage:       perPage,
		Status:        input.Status,
		CustomerEmail: input.CustomerEmail,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &OrderListResult{Orders: orders, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *OrderUsecase) GetOrder(ctx context.Context, id string) (*entity.Order, error) {
	return u.orderRepo.FindByID(ctx, id)
}

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := order.Status

	if err := u.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if oldStatus != status {
		order.Status = status
		// Notification failures shouldn't block the status update.
		_ = u.notificationUsecase.NotifyOrderStatusChanged(ctx, order, oldStatus, status)
	}
	return nil
}

func (u *OrderUsecase) UpdatePaymentStatus(ctx context.Context, id string, status entity.PaymentStatus) error {
	if _, err := u.orderRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.orderRepo.UpdatePaymentStatus(ctx, id, status)
}

type SalesReportResult struct {
	From    time.Time
	To      time.Time
	Summary *repository.SalesSummary
	Orders  []entity.Order
}

// GetSalesReport backs both the JSON preview (Summary) and the Excel/PDF
// export (Orders, the transaction-level ledger) from a single call, so the
// two callers (handler.GetSalesReport / handler.ExportSalesReport) never
// drift out of sync on what "the sales report for this range" means.
func (u *OrderUsecase) GetSalesReport(ctx context.Context, from, to time.Time) (*SalesReportResult, error) {
	summary, err := u.orderRepo.GetSalesSummary(ctx, from, to)
	if err != nil {
		return nil, err
	}

	orders, _, err := u.orderRepo.FindAll(ctx, repository.OrderFilter{
		Page:    1,
		PerPage: 10000,
		From:    &from,
		To:      &to,
	})
	if err != nil {
		return nil, err
	}

	return &SalesReportResult{From: from, To: to, Summary: summary, Orders: orders}, nil
}
