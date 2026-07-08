package response

import "github.com/labstack/echo/v4"

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func Success(c echo.Context, status int, message string, data interface{}) error {
	return c.JSON(status, envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessPaginated(c echo.Context, status int, data interface{}, meta Meta) error {
	return c.JSON(status, envelope{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c echo.Context, status int, message string, errs interface{}) error {
	return c.JSON(status, envelope{
		Success: false,
		Message: message,
		Errors:  errs,
	})
}
