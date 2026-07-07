package repository

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrAddressNotFound      = errors.New("address not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrProductNotFound      = errors.New("product not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrDuplicateEntry       = errors.New("duplicate entry")
)
