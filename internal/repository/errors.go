package repository

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrAddressNotFound        = errors.New("address not found")
	ErrRefreshTokenNotFound   = errors.New("refresh token not found")
	ErrProductNotFound        = errors.New("product not found")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrOrderNotFound          = errors.New("order not found")
	ErrServiceRequestNotFound = errors.New("service request not found")
	ErrReviewNotFound         = errors.New("review not found")
	ErrCouponNotFound         = errors.New("coupon not found")
	ErrSubscriberNotFound     = errors.New("subscriber not found")
	ErrSubscriberExists       = errors.New("email already subscribed")
	ErrCollectionNotFound         = errors.New("collection not found")
	ErrDuplicateEntry             = errors.New("duplicate entry")
	ErrVerificationTokenNotFound  = errors.New("verification token not found")
)
