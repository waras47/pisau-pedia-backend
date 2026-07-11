package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateOrderItemRequest struct {
	ProductSlug string `json:"product_slug" validate:"required"`
	Quantity    uint   `json:"quantity" validate:"required,min=1"`
}

type CreateOrderRequest struct {
	CustomerName       string                   `json:"customer_name" validate:"required,min=2"`
	CustomerEmail      string                   `json:"customer_email" validate:"required,email"`
	CustomerPhone      *string                  `json:"customer_phone"`
	ShippingAddress    string                   `json:"shipping_address" validate:"required"`
	ShippingCity       string                   `json:"shipping_city" validate:"required"`
	ShippingProvince   *string                  `json:"shipping_province"`
	ShippingPostalCode string                   `json:"shipping_postal_code" validate:"required"`
	CouponCode         string                   `json:"coupon_code"`
	Items              []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

func (r CreateOrderRequest) ToInput() usecase.CreateOrderInput {
	input := usecase.CreateOrderInput{
		CustomerName:       r.CustomerName,
		CustomerEmail:      r.CustomerEmail,
		CustomerPhone:      r.CustomerPhone,
		ShippingAddress:    r.ShippingAddress,
		ShippingCity:       r.ShippingCity,
		ShippingProvince:   r.ShippingProvince,
		ShippingPostalCode: r.ShippingPostalCode,
		CouponCode:         r.CouponCode,
	}
	for _, i := range r.Items {
		input.Items = append(input.Items, usecase.OrderItemInput{ProductSlug: i.ProductSlug, Quantity: i.Quantity})
	}
	return input
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing ready_for_delivery delivered cancelled"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" validate:"required,oneof=unpaid paid expired failed"`
}

type OrderItemResponse struct {
	ProductName string `json:"product_name"`
	ProductSlug string `json:"product_slug"`
	Price       int64  `json:"price"`
	Quantity    uint   `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
}

type OrderResponse struct {
	ID                 string              `json:"id"`
	Status             string              `json:"status"`
	PaymentStatus      string              `json:"payment_status"`
	CustomerName       string              `json:"customer_name"`
	CustomerEmail      string              `json:"customer_email"`
	CustomerPhone      string              `json:"customer_phone,omitempty"`
	ShippingAddress    string              `json:"shipping_address"`
	ShippingCity       string              `json:"shipping_city"`
	ShippingProvince   string              `json:"shipping_province,omitempty"`
	ShippingPostalCode string              `json:"shipping_postal_code"`
	Subtotal           int64               `json:"subtotal"`
	Total              int64               `json:"total"`
	Currency           string              `json:"currency"`
	CouponCode         string              `json:"coupon_code,omitempty"`
	DiscountAmount     int64               `json:"discount_amount,omitempty"`
	InvoiceURL         string              `json:"invoice_url,omitempty"`
	CreatedAt          string              `json:"created_at"`
	Items              []OrderItemResponse `json:"items,omitempty"`
}

func ToOrderResponse(o *entity.Order) OrderResponse {
	resp := OrderResponse{
		ID:                 o.ID,
		Status:             string(o.Status),
		PaymentStatus:      string(o.PaymentStatus),
		CustomerName:       o.CustomerName,
		CustomerEmail:      o.CustomerEmail,
		ShippingAddress:    o.ShippingAddress,
		ShippingCity:       o.ShippingCity,
		ShippingPostalCode: o.ShippingPostalCode,
		Subtotal:           o.Subtotal,
		Total:              o.Total,
		Currency:           o.Currency,
		DiscountAmount:     o.DiscountAmount,
		CreatedAt:          o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if o.CouponCode != nil {
		resp.CouponCode = *o.CouponCode
	}
	if o.CustomerPhone != nil {
		resp.CustomerPhone = *o.CustomerPhone
	}
	if o.ShippingProvince != nil {
		resp.ShippingProvince = *o.ShippingProvince
	}
	if o.XenditInvoiceURL != nil {
		resp.InvoiceURL = *o.XenditInvoiceURL
	}
	for _, i := range o.Items {
		resp.Items = append(resp.Items, OrderItemResponse{
			ProductName: i.ProductName,
			ProductSlug: i.ProductSlug,
			Price:       i.Price,
			Quantity:    i.Quantity,
			Subtotal:    i.Subtotal,
		})
	}
	return resp
}

func ToOrderResponses(orders []entity.Order) []OrderResponse {
	out := make([]OrderResponse, 0, len(orders))
	for i := range orders {
		out = append(out, ToOrderResponse(&orders[i]))
	}
	return out
}

type DailyRevenueResponse struct {
	Date    string `json:"date"`
	Revenue int64  `json:"revenue"`
}

type TopProductResponse struct {
	ProductName  string `json:"product_name"`
	QuantitySold int64  `json:"quantity_sold"`
	Revenue      int64  `json:"revenue"`
}

type SalesReportResponse struct {
	From         string                 `json:"from"`
	To           string                 `json:"to"`
	TotalRevenue int64                  `json:"total_revenue"`
	TotalOrders  int64                  `json:"total_orders"`
	StatusCounts map[string]int64       `json:"status_counts"`
	DailyRevenue []DailyRevenueResponse `json:"daily_revenue"`
	TopProducts  []TopProductResponse   `json:"top_products"`
}

func ToSalesReportResponse(r *usecase.SalesReportResult) SalesReportResponse {
	resp := SalesReportResponse{
		From:         r.From.Format("2006-01-02"),
		To:           r.To.Format("2006-01-02"),
		TotalRevenue: r.Summary.TotalRevenue,
		TotalOrders:  r.Summary.TotalOrders,
		StatusCounts: r.Summary.StatusCounts,
	}
	for _, d := range r.Summary.DailyRevenue {
		resp.DailyRevenue = append(resp.DailyRevenue, DailyRevenueResponse{Date: d.Date, Revenue: d.Revenue})
	}
	for _, p := range r.Summary.TopProducts {
		resp.TopProducts = append(resp.TopProducts, TopProductResponse{
			ProductName:  p.ProductName,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		})
	}
	return resp
}
