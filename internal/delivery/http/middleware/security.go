package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// Secure applies standard hardening headers. HSTS is only sent in
// production because it is meaningless (and confusing to browsers) over
// plain HTTP local development.
func Secure(isProduction bool) echo.MiddlewareFunc {
	hstsMaxAge := 0
	if isProduction {
		hstsMaxAge = 31536000 // 1 year
	}
	return echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            hstsMaxAge,
		HSTSExcludeSubdomains: !isProduction,
	})
}

// BodyLimit caps request body size so a single oversized request can't be
// used as a cheap denial-of-service vector.
func BodyLimit() echo.MiddlewareFunc {
	return echomw.BodyLimit("1M")
}
