package usecase

import "errors"

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountDisabled        = errors.New("account disabled")
	ErrInvalidRefreshToken    = errors.New("invalid or expired refresh token")
	ErrForbidden              = errors.New("forbidden")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
	ErrGoogleAuthUnavailable  = errors.New("google sign-in is not configured")
	ErrInvalidExchangeCode    = errors.New("invalid or expired exchange code")
	// ErrOrderNotEligibleForConfirmation covers every case ConfirmReceivedIfEligible's
	// guard rejects: not paid yet, still pending/cancelled — see
	// docs/16-plan-konfirmasi-pesanan-diterima-review.md.
	ErrOrderNotEligibleForConfirmation = errors.New("order is not eligible for receipt confirmation yet")
	ErrEmailNotVerified                = errors.New("email not verified")
	ErrVerificationTokenInvalid        = errors.New("invalid or expired verification token")
	ErrEmailAlreadyVerified            = errors.New("email already verified")
)
