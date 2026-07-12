package payment

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/komercepay"
)

// KomercePaymentGateway creates real VA/QRIS payments via the Komerce Payment
// Service and returns the hosted payment_url the customer is redirected to.
type KomercePaymentGateway struct {
	client   *komercepay.Client
	fallback *DummyGateway
}

func NewKomercePaymentGateway(client *komercepay.Client, fallback *DummyGateway) *KomercePaymentGateway {
	return &KomercePaymentGateway{client: client, fallback: fallback}
}

func (g *KomercePaymentGateway) CreateInvoice(ctx context.Context, input usecase.CreateInvoiceInput) (*usecase.PaymentInstruction, error) {
	// No credentials or no method chosen → fall back to the dummy success flow
	// so local dev without keys still works end-to-end.
	if !g.client.Enabled() || input.PaymentType == "" {
		return g.fallback.CreateInvoice(ctx, input)
	}

	items := make([]komercepay.CreateItem, 0, len(input.Items))
	for _, i := range input.Items {
		items = append(items, komercepay.CreateItem{Name: i.Name, Price: i.Price, Quantity: i.Quantity})
	}
	if len(items) == 0 {
		items = append(items, komercepay.CreateItem{Name: input.Description, Price: input.Amount, Quantity: 1})
	}

	data, err := g.client.CreatePayment(ctx, komercepay.CreateInput{
		OrderID:     input.ExternalID,
		PaymentType: input.PaymentType,
		ChannelCode: input.PaymentChannel,
		Amount:      input.Amount,
		Name:        input.CustomerName,
		Email:       input.CustomerEmail,
		Phone:       input.CustomerPhone,
		Items:       items,
	})
	if err != nil {
		return nil, err
	}

	return &usecase.PaymentInstruction{
		Provider:   "komerce",
		PaymentID:  data.PaymentID,
		PaymentURL: data.PaymentURL,
		VANumber:   data.VANumber,
		QRString:   data.QRString,
		ExpiryTime: data.ExpiredAt,
		InvoiceURL: data.PaymentURL,
	}, nil
}
