package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// AuthRateLimiter throttles brute-force attempts against login/register:
// 5 requests/second sustained with a burst of 10, tracked per client IP.
// Uses Echo's built-in in-memory store — no extra infrastructure (Redis)
// needed for this project's scale.
func AuthRateLimiter() echo.MiddlewareFunc {
	store := echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
		Rate:  rate.Limit(5),
		Burst: 10,
	})

	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusForbidden, "rate limiter error")
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "too many attempts, please try again later")
		},
	})
}
