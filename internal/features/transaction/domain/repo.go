package domain

import "context"

type BalanceRepo interface {
	GetBalance(ctx context.Context, userID int, currency string) (float64, error)
	UpdateBalance(ctx context.Context, userID int, currency string, amount float64) (float64, error)
}

type LedgerRepo interface {
	Create(ctx context.Context, entry *LedgerEntry) error
}

type WithdrawalRepo interface {
	Create(ctx context.Context, w *Withdrawal) error
	GetByID(ctx context.Context, withdrawalID int) (*Withdrawal, error)
	GetByIdempotencyKey(ctx context.Context, userID int, key string) (*Withdrawal, error)
	Update(ctx context.Context, w *Withdrawal) error
}
