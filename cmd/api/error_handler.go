package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

// httpErrorHandler prevents internal error details (SQL errors, stack
// traces, panics recovered by echomw.Recover) from leaking into API
// responses unless APP_DEBUG is explicitly enabled — information
// disclosure is a common stepping stone for attackers probing an API.
func httpErrorHandler(debug bool, log zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		message := "internal server error"

		if he, ok := err.(*echo.HTTPError); ok {
			status = he.Code
			if m, ok := he.Message.(string); ok {
				message = m
			}
		}

		if status == http.StatusInternalServerError {
			log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
			if debug {
				message = err.Error()
			}
		}

		if respErr := response.Error(c, status, message, nil); respErr != nil {
			log.Error().Err(respErr).Msg("failed to write error response")
		}
	}
}
