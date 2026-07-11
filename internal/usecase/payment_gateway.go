package usecase

import "context"

type CreateInvoiceInput struct {
	ExternalID    string
	Amount        int64
	CustomerName  string
	CustomerEmail string
	Description   string
}

type InvoiceResult struct {
	InvoiceURL string
}

// PaymentGateway abstracts the payment provider so OrderUsecase never talks
// to a specific vendor SDK directly — swap DummyGateway for a real Xendit
// (or other) implementation later without touching usecase/handler code.
type PaymentGateway interface {
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*InvoiceResult, error)
}
