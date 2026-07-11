package payment

import (
	"context"
	"fmt"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

// DummyGateway is a placeholder PaymentGateway used until a real provider
// (Xendit or otherwise) is wired in. It never charges anyone — it just
// points the customer straight at the success page. Orders created through
// it stay "unpaid" until an admin marks them paid manually in the admin
// panel (see OrderHandler.UpdatePaymentStatus).
type DummyGateway struct {
	frontendURL string
}

func NewDummyGateway(frontendURL string) *DummyGateway {
	return &DummyGateway{frontendURL: frontendURL}
}

func (g *DummyGateway) CreateInvoice(_ context.Context, input usecase.CreateInvoiceInput) (*usecase.InvoiceResult, error) {
	return &usecase.InvoiceResult{
		InvoiceURL: fmt.Sprintf("%s/checkout/success?order=%s", g.frontendURL, input.ExternalID),
	}, nil
}
