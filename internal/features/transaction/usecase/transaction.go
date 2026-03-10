package usecase

import (
	"context"
	"errors"
	"strconv"

	"back/internal/features/transaction/domain"

	"github.com/DangeL187/erax"
)

type txManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionUseCase struct {
	balanceRepo    domain.BalanceRepo
	ledgerRepo     domain.LedgerRepo
	withdrawalRepo domain.WithdrawalRepo
	txManager      txManager
}

func NewTransactionUseCase(br domain.BalanceRepo, lr domain.LedgerRepo, wr domain.WithdrawalRepo, txm txManager) *TransactionUseCase {
	return &TransactionUseCase{
		balanceRepo:    br,
		ledgerRepo:     lr,
		withdrawalRepo: wr,
		txManager:      txm,
	}
}

func (u *TransactionUseCase) ConfirmWithdrawal(ctx context.Context, userID int, withdrawalID string) error {
	withdrawalIdInt, err := strconv.Atoi(withdrawalID)
	if err != nil {
		return erax.Wrap(err, "failed to convert withdrawalID to int")
	}

	w, err := u.withdrawalRepo.GetByID(ctx, withdrawalIdInt)
	if err != nil {
		return erax.Wrap(err, "failed to get withdrawal by id")
	}
	if w.UserID != userID {
		return domain.ErrWithdrawalUserMismatch
	}

	w.Status = "confirmed"

	err = u.withdrawalRepo.Update(ctx, w)
	if err != nil {
		return erax.Wrap(err, "failed to update withdrawal")
	}

	return nil
}

// CreateWithdrawal performs an atomic withdrawal flow:
// 1) Checks for an existing withdrawal with the same idempotency key to prevent double spending.
// 2) Updates the user's balance.
// 3) Creates a new withdrawal record.
// 4) Creates a corresponding ledger entry.
// All steps are wrapped in a transaction.
func (u *TransactionUseCase) CreateWithdrawal(ctx context.Context, userID int, req *domain.CreateWithdrawalRequest) (*domain.Withdrawal, error) {
	var result *domain.Withdrawal

	err := u.txManager.WithinTx(ctx, func(ctx context.Context) error {
		existing, err := u.withdrawalRepo.GetByIdempotencyKey(ctx, userID, req.IdempotencyKey)
		if err == nil {
			if existing.Amount == req.Amount && existing.Currency == req.Currency &&
				existing.Destination == req.Destination {
				result = existing
				return nil
			}
			return domain.ErrIdempotencyConflict
		} else if !errors.Is(err, domain.ErrWithdrawalNotFound) {
			return erax.Wrap(err, "failed to get withdrawal by idempotency key")
		}

		_, err = u.balanceRepo.UpdateBalance(ctx, userID, req.Currency, -req.Amount)
		if err != nil {
			return erax.Wrap(err, "failed to update balance")
		}

		w := &domain.Withdrawal{
			UserID:         userID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Destination:    req.Destination,
			IdempotencyKey: req.IdempotencyKey,
			Status:         "pending",
		}

		err = u.withdrawalRepo.Create(ctx, w)
		if err != nil {
			return erax.Wrap(err, "failed to create withdrawal")
		}

		entry := &domain.LedgerEntry{
			UserID:        userID,
			ReferenceType: "withdrawal",
			ReferenceID:   w.ID,
		}

		err = u.ledgerRepo.Create(ctx, entry)
		if err != nil {
			return erax.Wrap(err, "failed to create ledger entry")
		}

		result = w
		return nil
	})

	if err != nil {
		return nil, erax.Wrap(err, "failed to perform transaction")
	}

	return result, nil
}

func (u *TransactionUseCase) GetWithdrawalByID(ctx context.Context, userID int, withdrawalID string) (*domain.Withdrawal, error) {
	withdrawalIdInt, err := strconv.Atoi(withdrawalID)
	if err != nil {
		return nil, erax.Wrap(err, "failed to convert withdrawalID to int")
	}

	w, err := u.withdrawalRepo.GetByID(ctx, withdrawalIdInt)
	if err != nil {
		return nil, erax.Wrap(err, "failed to get withdrawal by id")
	}
	if w.UserID != userID {
		return nil, domain.ErrWithdrawalUserMismatch
	}

	return w, nil
}
