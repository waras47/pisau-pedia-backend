package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/komercepay"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/mailer"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/rajaongkir"
)

var (
	ErrEmptyOrder        = errors.New("order must have at least one valid item")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type OrderItemInput struct {
	ProductSlug string
	Quantity    uint
}

type CreateOrderInput struct {
	// IdempotencyKey is a client-generated token unique per checkout attempt
	// (not per retry of that attempt). A double-submit — double-click,
	// network retry, multiple tabs — replays the same key, so CreateOrder
	// returns the order it already created instead of billing the customer
	// twice.
	IdempotencyKey string
	// UserID links the order to the account it was placed under, if any —
	// empty means guest checkout. Set only when the request carried a valid
	// access token (see appmw.OptionalJWTAuth); never trust a client-sent
	// value for this. Guest orders (UserID == "") never appear in "Pesanan
	// Saya" or become eligible for receipt confirmation — see
	// docs/16-plan-konfirmasi-pesanan-diterima-review.md.
	UserID             string
	CustomerName       string
	CustomerEmail      string
	CustomerPhone      *string
	ShippingAddress    string
	ShippingCity       string
	ShippingProvince   *string
	ShippingPostalCode string
	CouponCode         string
	DestinationID      string
	Courier            string
	Service            string
	PaymentType        string
	PaymentChannel     string
	Items              []OrderItemInput
	// IsManual marks an order the admin recorded by hand (e.g. a WhatsApp/
	// phone sale) purely for reporting — skips the RajaOngkir cost lookup
	// (ManualShippingCost is used verbatim instead) and the payment
	// gateway invoice, and is created already paid.
	IsManual           bool
	ManualShippingCost int64
}

type OrderListInput struct {
	Page          int
	PerPage       int
	Status        string
	CustomerEmail string
	Search        string
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
	paymentClient       *komercepay.Client
	mailer              *mailer.Mailer
	log                 zerolog.Logger
}

func NewOrderUsecase(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, paymentGateway PaymentGateway, couponUsecase *CouponUsecase, notificationUsecase *NotificationUsecase, shippingClient *rajaongkir.Client, paymentClient *komercepay.Client, m *mailer.Mailer, log zerolog.Logger) *OrderUsecase {
	return &OrderUsecase{orderRepo: orderRepo, productRepo: productRepo, paymentGateway: paymentGateway, couponUsecase: couponUsecase, notificationUsecase: notificationUsecase, shippingClient: shippingClient, paymentClient: paymentClient, mailer: m, log: log}
}

func strPtr(s string) *string { return &s }

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*entity.Order, error) {
	if input.IdempotencyKey != "" {
		existing, err := u.orderRepo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil && !errors.Is(err, repository.ErrOrderNotFound) {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	var items []entity.OrderItem
	var subtotal int64
	var totalWeight int

	// Prices and weight are recomputed from the product catalog, never trusted
	// from the client — the same defensive pattern the old Next.js checkout used.
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
		totalWeight += int(product.Weight) * int(i.Quantity)
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

	// Reserve stock up front, atomically per item, so two concurrent
	// checkouts can never both win the last unit — a customer who never pays
	// gets their reservation released automatically by ReleaseExpiredOrders,
	// and anything that fails downstream in this call releases immediately
	// via the deferred rollback below.
	var reserved []entity.OrderItem
	orderCommitted := false
	defer func() {
		if orderCommitted {
			return
		}
		for _, it := range reserved {
			_ = u.productRepo.RestoreStock(ctx, *it.ProductID, it.Quantity)
		}
	}()
	for _, it := range items {
		ok, err := u.productRepo.DecrementStockIfAvailable(ctx, *it.ProductID, it.Quantity)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInsufficientStock, it.ProductName)
		}
		reserved = append(reserved, it)
	}

	var couponCode *string
	var couponID string
	var discountAmount int64
	var freeShipping bool
	if input.CouponCode != "" {
		result, err := u.couponUsecase.ValidateCoupon(ctx, input.CouponCode, subtotal)
		if err != nil {
			return nil, err
		}
		couponCode = &input.CouponCode
		couponID = result.Coupon.ID
		discountAmount = result.DiscountAmount
		freeShipping = result.FreeShipping
	}

	// Shipping cost is recomputed server-side against RajaOngkir for the chosen
	// courier+service — the client-sent cost is never trusted. A free_shipping
	// coupon zeroes it out.
	var shippingCost int64
	var shippingCourier, shippingService, shippingETD, destinationID *string
	if input.IsManual {
		shippingCost = input.ManualShippingCost
	} else if input.DestinationID != "" && input.Courier != "" && input.Service != "" && u.shippingClient.Enabled() {
		if totalWeight < minWeightGrams {
			totalWeight = minWeightGrams
		}
		options, err := u.shippingClient.CalculateDomesticCost(ctx, u.shippingClient.OriginID(), input.DestinationID, totalWeight, []string{input.Courier})
		if err != nil {
			return nil, err
		}
		for _, opt := range options {
			if opt.Code == input.Courier && opt.Service == input.Service {
				shippingCost = opt.Cost
				etd := opt.ETD
				c, s, d := input.Courier, input.Service, input.DestinationID
				shippingCourier, shippingService, shippingETD, destinationID = &c, &s, &etd, &d
				break
			}
		}
	}
	if freeShipping {
		shippingCost = 0
	}

	orderID := uuid.New().String()
	now := time.Now()
	order := &entity.Order{
		ID:                 orderID,
		IdempotencyKey:     strPtrOrNil(input.IdempotencyKey),
		UserID:             strPtrOrNil(input.UserID),
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
		Total:              subtotal - discountAmount + shippingCost,
		Currency:           "IDR",
		CouponCode:         couponCode,
		DiscountAmount:     discountAmount,
		ShippingCost:       shippingCost,
		ShippingCourier:    shippingCourier,
		ShippingService:    shippingService,
		ShippingETD:        shippingETD,
		DestinationID:      destinationID,
		XenditExternalID:   &orderID,
		Items:              items,
	}

	if input.IsManual {
		// Recorded by an admin as already paid (e.g. a WhatsApp order
		// confirmed by bank transfer) — no invoice to generate.
		order.Status = entity.OrderStatusProcessing
		order.PaymentStatus = entity.PaymentStatusPaid
		order.PaymentProvider = strPtr("manual")
	} else {
		// Create the payment first so the instruction (payment_id, VA/QRIS,
		// hosted URL) is persisted in the same insert as the order.
		var customerPhone string
		if input.CustomerPhone != nil {
			customerPhone = *input.CustomerPhone
		}
		paymentItems := make([]PaymentItem, 0, len(items))
		for _, it := range items {
			paymentItems = append(paymentItems, PaymentItem{Name: it.ProductName, Price: it.Price, Quantity: it.Quantity})
		}

		instruction, err := u.paymentGateway.CreateInvoice(ctx, CreateInvoiceInput{
			ExternalID:     orderID,
			Amount:         order.Total,
			CustomerName:   order.CustomerName,
			CustomerEmail:  order.CustomerEmail,
			CustomerPhone:  customerPhone,
			Description:    fmt.Sprintf("Pisau Pedia order — %d item(s)", len(items)),
			PaymentType:    input.PaymentType,
			PaymentChannel: input.PaymentChannel,
			Items:          paymentItems,
		})
		if err != nil {
			return nil, err
		}

		order.XenditInvoiceURL = &instruction.InvoiceURL
		order.PaymentProvider = strPtr(instruction.Provider)
		order.PaymentID = strPtrOrNil(instruction.PaymentID)
		order.PaymentType = strPtrOrNil(input.PaymentType)
		order.PaymentChannel = strPtrOrNil(input.PaymentChannel)
		order.PaymentVANumber = strPtrOrNil(instruction.VANumber)
		order.PaymentQRString = strPtrOrNil(instruction.QRString)
		order.PaymentURL = strPtrOrNil(instruction.PaymentURL)
		order.PaymentExpiry = strPtrOrNil(instruction.ExpiryTime)
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		// A concurrent request with the same idempotency key won the race and
		// committed first — the UNIQUE index caught what the pre-check above
		// couldn't. That request made its own stock reservation, so this
		// one's reservation (still pending in the deferred rollback above)
		// must be released — return the order that actually exists instead
		// of failing the checkout.
		if errors.Is(err, repository.ErrDuplicateEntry) && input.IdempotencyKey != "" {
			return u.orderRepo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		}
		return nil, err
	}
	orderCommitted = true

	// From here on the order is real — it's persisted, its payment invoice
	// exists, its stock is reserved. Nothing past this point may fail the
	// request back to the customer as "checkout failed", because it isn't:
	// they'd retry (or the frontend's stored idempotency key would resubmit)
	// into a confusing "but I thought it failed" state despite the order
	// being perfectly valid. Failures here are logged, not returned.
	if couponID != "" {
		if err := u.couponUsecase.couponRepo.IncrementUsage(ctx, couponID); err != nil {
			u.log.Error().Err(err).Str("order_id", order.ID).Str("coupon_id", couponID).
				Msg("order created but failed to increment coupon usage count")
		}
	}

	if err := u.notificationUsecase.NotifyOrderCreated(ctx, order); err != nil {
		u.log.Warn().Err(err).Str("order_id", order.ID).Msg("order created but failed to send notification")
	}

	return order, nil
}

// ReleaseExpiredOrders finds still-unpaid orders past their payment
// deadline, cancels them, and releases the stock they reserved at checkout
// back to the catalog. Meant to be called periodically (see the ticker in
// cmd/api/main.go) — it's the automated counterpart to the admin's manual
// "Cek Status Pembayaran" button, and doubles as the safety net for it: a
// deadline passing doesn't necessarily mean nobody paid, it can mean the
// webhook that would have told us was lost, so every candidate is
// double-checked against Komerce directly before anything is cancelled.
func (u *OrderUsecase) ReleaseExpiredOrders(ctx context.Context) (int, error) {
	orders, err := u.orderRepo.FindExpiredUnpaidOrders(ctx, time.Now())
	if err != nil {
		return 0, err
	}

	released := 0
	for _, o := range orders {
		if u.reconcileIfActuallyPaid(ctx, o) {
			continue // webhook was lost, not the customer — order stays sold, stock stays reserved
		}

		// Conditional on still being unpaid: if a payment landed (webhook or
		// manual check) in the moment between the query above and this
		// update, this is a no-op and the order's stock stays reserved —
		// it was legitimately sold.
		changed, err := u.orderRepo.ExpireIfUnpaid(ctx, o.ID)
		if err != nil || !changed {
			continue
		}
		for _, item := range o.Items {
			if item.ProductID != nil {
				_ = u.productRepo.RestoreStock(ctx, *item.ProductID, item.Quantity)
			}
		}
		_ = u.orderRepo.UpdateStatus(ctx, o.ID, entity.OrderStatusCancelled)
		released++
	}
	return released, nil
}

// reconcileIfActuallyPaid asks Komerce directly whether an order Komerce's
// webhook never confirmed was in fact paid. If so it marks the order paid
// (same atomic MarkPaidIfUnpaid + notify path as the webhook and the manual
// check use) and returns true so the caller leaves it alone instead of
// expiring it. Any reason the check can't be made (dummy-gateway order,
// Komerce unreachable) falls through to false — the expiry sweep proceeds
// as if unconfirmed, same as it always has.
func (u *OrderUsecase) reconcileIfActuallyPaid(ctx context.Context, o entity.Order) bool {
	if u.paymentClient == nil || !u.paymentClient.Enabled() {
		return false
	}
	if o.PaymentProvider == nil || *o.PaymentProvider != "komerce" || o.PaymentID == nil || *o.PaymentID == "" {
		return false
	}

	status, err := u.paymentClient.GetStatus(ctx, *o.PaymentID)
	if err != nil || !strings.EqualFold(status.Status, "PAID") {
		return false
	}

	changed, err := u.orderRepo.MarkPaidIfUnpaid(ctx, o.ID)
	if err != nil {
		return false
	}
	if changed {
		o.PaymentStatus = entity.PaymentStatusPaid
		_ = u.notificationUsecase.NotifyOrderPaid(ctx, &o)
	}
	return true
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
		Search:        input.Search,
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

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, id string, status entity.OrderStatus, trackingNumber, shippingEvidenceURL *string) error {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := order.Status

	if err := u.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if trackingNumber != nil || shippingEvidenceURL != nil {
		if err := u.orderRepo.UpdateShippingEvidence(ctx, id, trackingNumber, shippingEvidenceURL); err != nil {
			return err
		}
	}

	if oldStatus != status {
		order.Status = status
		_ = u.notificationUsecase.NotifyOrderStatusChanged(ctx, order, oldStatus, status)

		if u.mailer != nil && u.mailer.Enabled() && order.CustomerEmail != "" {
			go func() {
				if err := u.mailer.SendOrderStatusEmail(
					order.CustomerEmail, order.CustomerName, order.ID,
					string(oldStatus), string(status),
				); err != nil {
					u.log.Error().Err(err).Str("order_id", order.ID).Msg("failed to send order status email")
				}
			}()
		}
	}
	return nil
}

func (u *OrderUsecase) UpdatePaymentStatus(ctx context.Context, id string, status entity.PaymentStatus) error {
	if _, err := u.orderRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.orderRepo.UpdatePaymentStatus(ctx, id, status)
}

// UploadPaymentProof records the URL of a customer-uploaded receipt against
// their order — part of the manual/static payment flow (see DummyGateway).
// It doesn't change PaymentStatus; an admin still confirms payment manually
// after checking the proof, via UpdatePaymentStatus.
func (u *OrderUsecase) UploadPaymentProof(ctx context.Context, id string, url string) (*entity.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.orderRepo.UpdatePaymentProof(ctx, id, url); err != nil {
		return nil, err
	}
	order.PaymentProofURL = &url
	return order, nil
}

// ListMyOrders lists orders placed by userID while signed in — guest orders
// never appear here even if the email matches an account created later, by
// design (see docs/16-plan-konfirmasi-pesanan-diterima-review.md).
func (u *OrderUsecase) ListMyOrders(ctx context.Context, userID string, page, perPage int) (*OrderListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	orders, total, err := u.orderRepo.FindAll(ctx, repository.OrderFilter{
		Page:    page,
		PerPage: perPage,
		UserID:  userID,
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

// GetMyOrder returns an order only if it belongs to userID. A mismatch (or
// a guest order with no user_id at all) returns ErrOrderNotFound rather
// than a distinct "forbidden" error, so the response never confirms
// whether the order ID exists.
func (u *OrderUsecase) GetMyOrder(ctx context.Context, userID, orderID string) (*entity.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID == nil || *order.UserID != userID {
		return nil, repository.ErrOrderNotFound
	}
	return order, nil
}

// ConfirmReceived lets a customer self-report their package arrived. It
// deliberately never touches Status — that stays under admin control; see
// docs/16-plan-konfirmasi-pesanan-diterima-review.md for why.
func (u *OrderUsecase) ConfirmReceived(ctx context.Context, userID, orderID string) (*entity.Order, error) {
	order, err := u.GetMyOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}

	changed, err := u.orderRepo.ConfirmReceivedIfEligible(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	if !changed {
		if order.CustomerConfirmedAt != nil {
			return order, nil // already confirmed earlier — clicking again is a harmless no-op
		}
		return nil, ErrOrderNotEligibleForConfirmation
	}

	now := time.Now()
	order.CustomerConfirmedAt = &now
	_ = u.notificationUsecase.NotifyOrderReceived(ctx, order)
	return order, nil
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
