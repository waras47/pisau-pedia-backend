package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/komercepay"
)

var (
	ErrInvalidCallbackSignature = errors.New("invalid callback signature")
	ErrPaymentUnavailable       = errors.New("payment service is not configured")
	ErrNoPaymentToCheck         = errors.New("order has no komerce payment to check")
)

type PaymentUsecase struct {
	client              *komercepay.Client
	orderRepo           repository.OrderRepository
	notificationUsecase *NotificationUsecase
	callbackKey         string
}

func NewPaymentUsecase(client *komercepay.Client, orderRepo repository.OrderRepository, notificationUsecase *NotificationUsecase, callbackKey string) *PaymentUsecase {
	return &PaymentUsecase{client: client, orderRepo: orderRepo, notificationUsecase: notificationUsecase, callbackKey: callbackKey}
}

func (u *PaymentUsecase) GetMethods(ctx context.Context) ([]komercepay.PaymentMethod, error) {
	if !u.client.Enabled() {
		return nil, ErrPaymentUnavailable
	}
	return u.client.GetMethods(ctx)
}

// callbackPayload captures the fields we rely on from the Komerce webhook. The
// callback mirrors the payment-status shape (order_id + status).
type callbackPayload struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// HandleKomerceCallback verifies the HMAC-SHA256 signature over the raw body
// and, if the payment is PAID, marks the matching order paid.
func (u *PaymentUsecase) HandleKomerceCallback(ctx context.Context, rawBody []byte, signature string) error {
	if !u.verifySignature(rawBody, signature) {
		return ErrInvalidCallbackSignature
	}

	var payload callbackPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return err
	}
	if payload.OrderID == "" || !strings.EqualFold(payload.Status, "PAID") {
		return nil // nothing to do for non-paid states
	}

	order, err := u.orderRepo.FindByID(ctx, payload.OrderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil
		}
		return err
	}

	// Atomic conditional update instead of check-then-act: if two payment
	// confirmations for the same order race each other (Komerce retries
	// webhook delivery until it gets a 200, or a webhook lands at the same
	// time as an admin's manual status check), only the delivery that
	// actually flips unpaid->paid fires the notification.
	changed, err := u.orderRepo.MarkPaidIfUnpaid(ctx, order.ID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	order.PaymentStatus = entity.PaymentStatusPaid
	_ = u.notificationUsecase.NotifyOrderPaid(ctx, order)
	return nil
}

// CheckStatus asks Komerce directly for the current status of an order's
// payment instead of waiting for a webhook — the fallback for local
// development (Komerce can't reach a localhost webhook URL) and for
// production cases where a webhook delivery was missed entirely (e.g. the
// server was down when Komerce tried to call it).
func (u *PaymentUsecase) CheckStatus(ctx context.Context, orderID string) (*entity.Order, error) {
	if !u.client.Enabled() {
		return nil, ErrPaymentUnavailable
	}

	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.PaymentProvider == nil || *order.PaymentProvider != "komerce" || order.PaymentID == nil || *order.PaymentID == "" {
		return nil, ErrNoPaymentToCheck
	}
	if order.PaymentStatus == entity.PaymentStatusPaid {
		return order, nil // already settled, nothing to reconcile
	}

	status, err := u.client.GetStatus(ctx, *order.PaymentID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(status.Status, "PAID") {
		return order, nil
	}

	changed, err := u.orderRepo.MarkPaidIfUnpaid(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	if changed {
		order.PaymentStatus = entity.PaymentStatusPaid
		_ = u.notificationUsecase.NotifyOrderPaid(ctx, order)
	}
	return order, nil
}

func (u *PaymentUsecase) verifySignature(rawBody []byte, signature string) bool {
	if u.callbackKey == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(u.callbackKey))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}
