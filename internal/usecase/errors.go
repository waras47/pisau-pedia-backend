package usecase

import "errors"

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountDisabled        = errors.New("account disabled")
	ErrInvalidRefreshToken    = errors.New("invalid or expired refresh token")
	ErrForbidden              = errors.New("forbidden")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
)
