package entity

import "time"

type OrderStatus string

const (
	OrderStatusPending          OrderStatus = "pending"
	OrderStatusProcessing       OrderStatus = "processing"
	OrderStatusReadyForDelivery OrderStatus = "ready_for_delivery"
	OrderStatusDelivered        OrderStatus = "delivered"
	OrderStatusCancelled        OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "unpaid"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusExpired PaymentStatus = "expired"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type Order struct {
	ID                 string        `db:"id"`
	IdempotencyKey     *string       `db:"idempotency_key"`
	UserID             *string       `db:"user_id"`
	Status             OrderStatus   `db:"status"`
	PaymentStatus      PaymentStatus `db:"payment_status"`
	PaymentProvider    *string       `db:"payment_provider"`
	PaymentID          *string       `db:"payment_id"`
	PaymentType        *string       `db:"payment_type"`
	PaymentChannel     *string       `db:"payment_channel"`
	PaymentVANumber    *string       `db:"payment_va_number"`
	PaymentQRString    *string       `db:"payment_qr_string"`
	PaymentURL         *string       `db:"payment_url"`
	PaymentExpiry      *string       `db:"payment_expiry"`
	CustomerName       string        `db:"customer_name"`
	CustomerEmail      string        `db:"customer_email"`
	CustomerPhone      *string       `db:"customer_phone"`
	ShippingAddress    string        `db:"shipping_address"`
	ShippingCity       string        `db:"shipping_city"`
	ShippingProvince   *string       `db:"shipping_province"`
	ShippingPostalCode string        `db:"shipping_postal_code"`
	Subtotal           int64         `db:"subtotal"`
	Total              int64         `db:"total"`
	Currency           string        `db:"currency"`
	CouponCode         *string       `db:"coupon_code"`
	DiscountAmount     int64         `db:"discount_amount"`
	ShippingCost       int64         `db:"shipping_cost"`
	ShippingCourier    *string       `db:"shipping_courier"`
	ShippingService    *string       `db:"shipping_service"`
	ShippingETD        *string       `db:"shipping_etd"`
	DestinationID      *string       `db:"destination_id"`
	XenditExternalID   *string       `db:"xendit_external_id"`
	XenditInvoiceURL   *string       `db:"xendit_invoice_url"`
	CreatedAt          time.Time     `db:"created_at"`
	UpdatedAt          time.Time     `db:"updated_at"`

	Items []OrderItem `db:"-"`
}

type OrderItem struct {
	ID          string  `db:"id"`
	OrderID     string  `db:"order_id"`
	ProductID   *string `db:"product_id"`
	ProductName string  `db:"product_name"`
	ProductSlug string  `db:"product_slug"`
	Price       int64   `db:"price"`
	Quantity    uint    `db:"quantity"`
	Subtotal    int64   `db:"subtotal"`
}
