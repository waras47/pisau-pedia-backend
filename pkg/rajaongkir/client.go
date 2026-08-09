// Package rajaongkir wraps the RajaOngkir (Komerce) V2 shipping-cost API:
// destination search + domestic/international cost calculation. The API key
// is passed via the "key" header on every request.
package rajaongkir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	originID   string
	httpClient *http.Client
}

func New(baseURL, apiKey, originID string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		originID:   originID,
		// The combined multi-courier cost query is occasionally slow on
		// Komerce's sandbox backend (observed up to ~15s even after the
		// IPv4 fix below) — give it real headroom instead of failing
		// checkout over transient sandbox latency.
		httpClient: &http.Client{Timeout: 25 * time.Second, Transport: ipv4OnlyTransport()},
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

// OriginID returns the configured store origin location id.
func (c *Client) OriginID() string { return c.originID }

// Enabled reports whether the client has the minimum config to make calls.
func (c *Client) Enabled() bool { return c.apiKey != "" }

// Destination is a searchable area (subdistrict-level) returned by search.
type Destination struct {
	ID              int64  `json:"id"`
	Label           string `json:"label"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	ZipCode         string `json:"zip_code"`
}

// ShippingOption is one courier service quote.
type ShippingOption struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	ETD         string `json:"etd"`
}

type metaEnvelope struct {
	Meta struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
	} `json:"meta"`
}

type destinationResponse struct {
	metaEnvelope
	Data []Destination `json:"data"`
}

type internationalDestinationRaw struct {
	CountryID   string `json:"country_id"`
	CountryName string `json:"country_name"`
}

type internationalDestinationResponse struct {
	metaEnvelope
	Data []internationalDestinationRaw `json:"data"`
}

type costResponse struct {
	metaEnvelope
	Data []ShippingOption `json:"data"`
}

func (c *Client) SearchDomesticDestination(ctx context.Context, search string, limit int) ([]Destination, error) {
	return c.searchDestination(ctx, "/destination/domestic-destination", search, limit)
}

func (c *Client) SearchInternationalDestination(ctx context.Context, search string, limit int) ([]Destination, error) {
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("search", search)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", "0")

	path := "/destination/international-destination"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return []Destination{}, nil
	}

	var body internationalDestinationResponse
	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rajaongkir international destination search returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}

	destinations := make([]Destination, 0, len(body.Data))
	for _, raw := range body.Data {
		id, _ := strconv.ParseInt(raw.CountryID, 10, 64)
		destinations = append(destinations, Destination{
			ID:    id,
			Label: raw.CountryName,
			CityName: raw.CountryName,
		})
	}
	return destinations, nil
}

func (c *Client) searchDestination(ctx context.Context, path, search string, limit int) ([]Destination, error) {
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("search", search)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", "0")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var body destinationResponse
	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(&body); err != nil {
		return nil, err
	}
	// 404 = no match; treat as an empty result rather than an error.
	if resp.StatusCode == http.StatusNotFound {
		return []Destination{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rajaongkir destination search returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}
	return body.Data, nil
}

func (c *Client) CalculateDomesticCost(ctx context.Context, originID, destinationID string, weightGrams int, couriers []string) ([]ShippingOption, error) {
	return c.calculateCost(ctx, "/calculate/domestic-cost", originID, destinationID, weightGrams, couriers)
}

func (c *Client) CalculateInternationalCost(ctx context.Context, originID, destinationID string, weightGrams int, couriers []string) ([]ShippingOption, error) {
	return c.calculateCost(ctx, "/calculate/international-cost", originID, destinationID, weightGrams, couriers)
}

func (c *Client) calculateCost(ctx context.Context, path, originID, destinationID string, weightGrams int, couriers []string) ([]ShippingOption, error) {
	if weightGrams < 1 {
		weightGrams = 1
	}
	form := url.Values{}
	form.Set("origin", originID)
	form.Set("destination", destinationID)
	form.Set("weight", strconv.Itoa(weightGrams))
	form.Set("courier", strings.Join(couriers, ":"))
	form.Set("price", "lowest")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("key", c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body costResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rajaongkir cost calc returned status %d: %s", resp.StatusCode, body.Meta.Message)
	}
	return body.Data, nil
}
