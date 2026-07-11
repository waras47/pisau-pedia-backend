package exchangerate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	sourceURL = "https://open.er-api.com/v6/latest/USD"
	cacheTTL  = 6 * time.Hour
)

// apiResponse mirrors the subset of open.er-api.com's response we care about.
type apiResponse struct {
	Result      string             `json:"result"`
	Rates       map[string]float64 `json:"rates"`
	TimeLastUTC string             `json:"time_last_update_utc"`
}

// Snapshot is the cached exchange rate table, base USD (1 USD = Rates[CUR]).
type Snapshot struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// Client fetches USD-based exchange rates from a free public API and caches
// them in memory, so repeated frontend requests don't each trigger an
// outbound call — the upstream API has no key but is rate-limited per IP.
type Client struct {
	httpClient *http.Client

	mu       sync.RWMutex
	cached   *Snapshot
	fetchErr error
}

func New() *Client {
	return &Client{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// GetRates returns the cached snapshot if still fresh, otherwise fetches a
// new one. On fetch failure it falls back to a stale cached snapshot (if any)
// rather than failing the request outright.
func (c *Client) GetRates(ctx context.Context) (*Snapshot, error) {
	c.mu.RLock()
	fresh := c.cached != nil && time.Since(c.cached.UpdatedAt) < cacheTTL
	snapshot := c.cached
	c.mu.RUnlock()

	if fresh {
		return snapshot, nil
	}

	newSnapshot, err := c.fetch(ctx)
	if err != nil {
		if snapshot != nil {
			return snapshot, nil
		}
		return nil, err
	}

	c.mu.Lock()
	c.cached = newSnapshot
	c.mu.Unlock()

	return newSnapshot, nil
}

func (c *Client) fetch(ctx context.Context) (*Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange rate API returned status %d", resp.StatusCode)
	}

	var body apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Result != "success" {
		return nil, fmt.Errorf("exchange rate API result: %s", body.Result)
	}

	return &Snapshot{
		Base:      "USD",
		Rates:     body.Rates,
		UpdatedAt: time.Now(),
	}, nil
}
