package usecase

import "context"

type PaymentItem struct {
	Name     string
	Price    int64
	Quantity uint
}

type CreateInvoiceInput struct {
	ExternalID    string
	Amount        int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	Description   string
	// PaymentType is "bank_transfer" or "qris"; PaymentChannel is the VA bank
	// code (e.g. "BCA") when PaymentType is bank_transfer.
	PaymentType    string
	PaymentChannel string
	Items          []PaymentItem
}

// PaymentInstruction is what the customer needs to complete payment. For the
// Komerce gateway, PaymentURL is a hosted page that renders the VA/QRIS.
type PaymentInstruction struct {
	Provider   string
	PaymentID  string
	PaymentURL string
	VANumber   string
	QRString   string
	ExpiryTime string
	// InvoiceURL is where the frontend redirects the customer to pay.
	InvoiceURL string
}

// PaymentGateway abstracts the payment provider so OrderUsecase never talks to
// a specific vendor SDK directly — DummyGateway is used when no real provider
// is configured (dev), KomercePaymentGateway when credentials are present.
type PaymentGateway interface {
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*PaymentInstruction, error)
}
