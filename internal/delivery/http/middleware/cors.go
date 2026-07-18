package middleware

import (
	"net/url"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS always allows the configured frontend URL. In non-production, it
// also allows any http://localhost:<port> or http://127.0.0.1:<port>
// origin — Next.js auto-increments past its default port whenever
// something else already holds it (a second `npm run dev`, an unrelated
// local process), which happens often enough here that hardcoding a
// single dev port breaks login/API calls every time it shifts.
func CORS(frontendURL string, isDev bool) echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			if origin == frontendURL {
				return true, nil
			}
			if isDev {
				u, err := url.Parse(origin)
				if err == nil && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1") {
					return true, nil
				}
			}
			return false, nil
		},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE"},
	})
}
