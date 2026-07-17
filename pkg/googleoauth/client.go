// Package googleoauth wraps Google's OAuth 2.0 "Sign in with Google" flow:
// build the consent-screen URL, exchange the returned authorization code for
// the user's verified email/name.
package googleoauth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ErrEmailNotVerified = errors.New("google account email is not verified")

type Client struct {
	cfg        *oauth2.Config
	httpClient *http.Client
}

func New(clientID, clientSecret, redirectURL string) *Client {
	return &Client{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		httpClient: &http.Client{Timeout: 15 * time.Second, Transport: ipv4OnlyTransport()},
	}
}

// Enabled reports whether the client has the minimum config to make calls —
// Google Sign-In is off (button hidden, endpoints 503) until real
// credentials are set.
func (c *Client) Enabled() bool { return c.cfg.ClientID != "" && c.cfg.ClientSecret != "" }

// AuthCodeURL returns the URL to send the browser to for Google's consent
// screen. state must be an unpredictable, per-request value verified on
// callback (CSRF protection) — see AuthHandler.GoogleLogin.
func (c *Client) AuthCodeURL(state string) string {
	return c.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// UserInfo is the subset of Google's OIDC userinfo response this app needs.
type UserInfo struct {
	Sub           string `json:"sub"` // stable per-account identifier — stored as users.google_id
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Exchange trades a one-time authorization code (from the callback redirect)
// for the signed-in user's profile.
func (c *Client) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)

	token, err := c.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("google userinfo request failed")
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if !info.EmailVerified {
		return nil, ErrEmailNotVerified
	}
	return &info, nil
}

// ipv4OnlyTransport forces outbound connections over IPv4. On this project's
// dev network, IPv6 to Google/Cloudflare-fronted APIs completes the TCP
// handshake but then hangs indefinitely on the request itself (see the same
// fix in pkg/rajaongkir and pkg/komercepay) — Go's normal dual-stack dialing
// doesn't fail fast enough to fall back to IPv4 on its own.
func ipv4OnlyTransport() *http.Transport {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
	}
}
