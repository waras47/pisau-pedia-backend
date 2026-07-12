package dto

import "github.com/pisaupediaprojek/pisau-pedia-backend/pkg/komercepay"

type PaymentMethodResponse struct {
	PaymentType string `json:"payment_type"`
	DisplayName string `json:"display_name"`
	BankCode    string `json:"bank_code"`
	LogoURL     string `json:"logo_url"`
	MinAmount   int64  `json:"min_amount"`
	MaxAmount   int64  `json:"max_amount"`
}

func ToPaymentMethodResponses(methods []komercepay.PaymentMethod) []PaymentMethodResponse {
	out := make([]PaymentMethodResponse, 0, len(methods))
	for _, m := range methods {
		out = append(out, PaymentMethodResponse{
			PaymentType: m.PaymentType,
			DisplayName: m.DisplayName,
			BankCode:    m.BankCode,
			LogoURL:     m.LogoURL,
			MinAmount:   m.MinAmount,
			MaxAmount:   m.MaxAmount,
		})
	}
	return out
}
