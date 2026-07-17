package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyRole   = "role"
)

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed authorization header")
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			userID, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)
			if userID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
			}

			c.Set(ContextKeyUserID, userID)
			c.Set(ContextKeyRole, role)
			return next(c)
		}
	}
}

// OptionalJWTAuth is JWTAuth's non-enforcing sibling: a valid token
// populates the same context keys, but a missing/invalid one just proceeds
// as an anonymous request instead of rejecting it. Used on public endpoints
// that behave differently for a signed-in caller without requiring login —
// order creation is the motivating case (guest checkout stays allowed, but
// a logged-in customer's orders get user_id set so "Pesanan Saya" and
// receipt confirmation can find them later; see
// docs/16-plan-konfirmasi-pesanan-diterima-review.md).
func OptionalJWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				return next(c)
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return next(c)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return next(c)
			}
			if userID, _ := claims["sub"].(string); userID != "" {
				c.Set(ContextKeyUserID, userID)
				role, _ := claims["role"].(string)
				c.Set(ContextKeyRole, role)
			}
			return next(c)
		}
	}
}

func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			current, _ := c.Get(ContextKeyRole).(string)
			if current != role {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}
			return next(c)
		}
	}
}
