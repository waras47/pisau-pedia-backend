// Package komercepay wraps the Komerce (RajaOngkir) Payment Service API:
// create VA/QRIS payments, check status, cancel. Auth is the x-api-key header.
package komercepay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 20 * time.Second, Transport: ipv4OnlyTransport()},
	}
}

// ipv4OnlyTransport forces outbound connections over IPv4. On some networks
// the Cloudflare-fronted Komerce API completes an IPv6 TCP handshake but
// then hangs indefinitely on the request itself, so Go's normal
// dual-stack/Happy-Eyeballs dialing doesn't fail fast enough to fall back —
// every call silently eats the full client timeout instead.
func ipv4OnlyTransport() *http.Transport {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
	}
}

func (c *Client) Enabled() bool { return c.apiKey != "" }

type meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

type PaymentMethod struct {
	PaymentType string `json:"payment_type"`
	DisplayName string `json:"display_name"`
	BankCode    string `json:"bank_code"`
	LogoURL     string `json:"logo_url"`
	MinAmount   int64  `json:"min_amount"`
	MaxAmount   int64  `json:"max_amount"`
	Currency    string `json:"currency"`
}

type methodsResponse struct {
	Meta meta            `json:"meta"`
	Data []PaymentMethod `json:"data"`
}

func (c *Client) GetMethods(ctx context.Context) ([]PaymentMethod, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/user/methods", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body methodsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("komerce methods returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}
	return body.Data, nil
}

type CreateItem struct {
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity uint   `json:"quantity"`
}

type CreateInput struct {
	OrderID     string
	PaymentType string // "bank_transfer" | "qris"
	ChannelCode string // bank code for VA (e.g. "BCA"); empty for QRIS
	Amount      int64
	Name        string
	Email       string
	Phone       string
	Items       []CreateItem
}

type PaymentData struct {
	PaymentID  string `json:"payment_id"`
	ExternalID string `json:"external_id"`
	PaymentURL string `json:"payment_url"`
	VANumber   string `json:"va_number"`
	QRString   string `json:"qr_string"`
	BankCode   string `json:"bank_code"`
	BankName   string `json:"bank_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	ExpiredAt  string `json:"expired_at"`
	CreatedAt  string `json:"created_at"`
}

type createResponse struct {
	Meta meta        `json:"meta"`
	Data PaymentData `json:"data"`
}

func (c *Client) CreatePayment(ctx context.Context, input CreateInput) (*PaymentData, error) {
	phone := input.Phone
	if phone == "" {
		phone = "0000000000"
	}
	payload := map[string]interface{}{
		"order_id":     input.OrderID,
		"payment_type": input.PaymentType,
		"amount":       input.Amount,
		"customer": map[string]string{
			"name":  input.Name,
			"email": input.Email,
			"phone": phone,
		},
		"items": input.Items,
	}
	if input.ChannelCode != "" {
		payload["channel_code"] = input.ChannelCode
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/user/payment/create", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body createResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("komerce create payment returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}
	return &body.Data, nil
}

type StatusData struct {
	PaymentID   string `json:"payment_id"`
	OrderID     string `json:"order_id"`
	BankCode    string `json:"bank_code"`
	VANumber    string `json:"va_number"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	PaymentType string `json:"payment_type"`
	Status      string `json:"status"`
	SenderName  string `json:"sender_name"`
	ExpiredAt   string `json:"expired_at"`
	CreatedAt   string `json:"created_at"`
	PaidAt      string `json:"paid_at"`
}

type statusResponse struct {
	Meta meta       `json:"meta"`
	Data StatusData `json:"data"`
}

func (c *Client) GetStatus(ctx context.Context, paymentID string) (*StatusData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/user/payment/status/"+paymentID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("komerce payment status returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}
	return &body.Data, nil
}
