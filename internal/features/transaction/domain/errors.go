package domain

import "errors"

var (
	ErrIdempotencyConflict    = errors.New("idempotency key conflict")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrWithdrawalNotFound     = errors.New("withdrawal not found")
	ErrWithdrawalUserMismatch = errors.New("withdrawal belongs to another user")
)
