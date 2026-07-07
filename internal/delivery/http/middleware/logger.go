package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func RequestLogger(log zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			log.Info().
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Int("status", c.Response().Status).
				Dur("latency", time.Since(start)).
				Msg("request")

			return err
		}
	}
}
