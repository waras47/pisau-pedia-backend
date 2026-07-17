package handler

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	frontendURL string
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase, frontendURL string) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, frontendURL: frontendURL}
}

// googleStateCookie carries the CSRF state across the redirect to Google and
// back — short-lived, httpOnly, and scoped to the callback path only.
const googleStateCookie = "google_oauth_state"

func randomState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf), nil
}

func isSecureRequest(c echo.Context) bool {
	return c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, tokens, err := h.authUsecase.Register(c.Request().Context(), req.Email, req.Password, req.FullName, req.Phone)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
			return response.Error(c, http.StatusConflict, "email already registered", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to register", nil)
	}

	return response.Success(c, http.StatusCreated, "Registration successful", dto.ToAuthResponse(user, tokens))
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, tokens, err := h.authUsecase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		case errors.Is(err, usecase.ErrAccountDisabled):
			return response.Error(c, http.StatusForbidden, "account disabled", nil)
		default:
			return response.Error(c, http.StatusInternalServerError, "failed to login", nil)
		}
	}

	return response.Success(c, http.StatusOK, "Login successful", dto.ToAuthResponse(user, tokens))
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	tokens, err := h.authUsecase.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token", nil)
	}

	return response.Success(c, http.StatusOK, "Token refreshed", map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req dto.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	if err := h.authUsecase.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to logout", nil)
	}

	return response.Success(c, http.StatusOK, "Logged out", nil)
}

// GoogleStatus lets the frontend decide whether to render the "Continue
// with Google" button at all — hidden rather than shown-then-broken when no
// credentials are configured.
func (h *AuthHandler) GoogleStatus(c echo.Context) error {
	return response.Success(c, http.StatusOK, "OK", map[string]bool{"enabled": h.authUsecase.GoogleEnabled()})
}

// GoogleLogin starts the flow: stash a CSRF state in a short-lived cookie,
// then redirect the browser to Google's consent screen.
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	state, err := randomState()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to start google login", nil)
	}

	authURL, err := h.authUsecase.GoogleAuthURL(state)
	if err != nil {
		if errors.Is(err, usecase.ErrGoogleAuthUnavailable) {
			return response.Error(c, http.StatusServiceUnavailable, "google sign-in is not configured", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to start google login", nil)
	}

	c.SetCookie(&http.Cookie{
		Name:     googleStateCookie,
		Value:    state,
		Path:     "/api/v1/auth/google",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback is where Google redirects back with an authorization code.
// It verifies the CSRF state, completes the login server-side, then hands
// the browser off to the frontend with a single-use exchange code — never
// the real tokens — in the URL.
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	failureURL := h.frontendURL + "/account/login?error=google_failed"

	cookie, err := c.Cookie(googleStateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != c.QueryParam("state") {
		return c.Redirect(http.StatusFound, failureURL)
	}
	c.SetCookie(&http.Cookie{
		Name: googleStateCookie, Value: "", Path: "/api/v1/auth/google", MaxAge: -1,
	})

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusFound, failureURL)
	}

	user, tokens, err := h.authUsecase.GoogleCallback(c.Request().Context(), code)
	if err != nil {
		return c.Redirect(http.StatusFound, failureURL)
	}

	exchangeCode, err := h.authUsecase.IssueExchangeCode(user, tokens)
	if err != nil {
		return c.Redirect(http.StatusFound, failureURL)
	}

	return c.Redirect(http.StatusFound, h.frontendURL+"/account/callback?code="+exchangeCode)
}

// GoogleExchange is called by the frontend's callback page immediately
// after landing there — trades the one-time code from the redirect for the
// real access/refresh tokens, same response shape as Login/Register.
func (h *AuthHandler) GoogleExchange(c echo.Context) error {
	var req dto.GoogleExchangeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	user, tokens, err := h.authUsecase.ConsumeExchangeCode(req.Code)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid or expired code", nil)
	}

	return response.Success(c, http.StatusOK, "Login successful", dto.ToAuthResponse(user, tokens))
}
