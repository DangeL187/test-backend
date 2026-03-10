package infra

import (
	"context"
	"errors"

	"github.com/DangeL187/erax"
	"gorm.io/gorm"

	"back/internal/features/transaction/domain"
)

type WithdrawalRepo struct {
	db *gorm.DB
}

func NewWithdrawalRepo(db *gorm.DB) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

func (r *WithdrawalRepo) Create(ctx context.Context, w *domain.Withdrawal) error {
	err := r.db.WithContext(ctx).Create(w).Error
	if err != nil {
		err = erax.Wrap(err, "failed to create withdrawal")
		return erax.WithMeta(err, "layer", "DB")
	}

	return nil
}

func (r *WithdrawalRepo) GetByID(ctx context.Context, withdrawalID int) (*domain.Withdrawal, error) {
	var w domain.Withdrawal
	err := r.db.WithContext(ctx).First(&w, withdrawalID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrWithdrawalNotFound
	} else if err != nil {
		err = erax.Wrap(err, "failed to get withdrawal by id")
		return nil, erax.WithMeta(err, "layer", "DB")
	}

	return &w, nil
}

func (r *WithdrawalRepo) GetByIdempotencyKey(ctx context.Context, userID int, key string) (*domain.Withdrawal, error) {
	var w domain.Withdrawal
	err := r.db.WithContext(ctx).Where("user_id = ? AND idempotency_key = ?", userID, key).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrWithdrawalNotFound
	} else if err != nil {
		err = erax.Wrap(err, "failed to get withdrawal by idempotency_key")
		return nil, erax.WithMeta(err, "layer", "DB")
	}

	return &w, nil
}

func (r *WithdrawalRepo) Update(ctx context.Context, w *domain.Withdrawal) error {
	err := r.db.WithContext(ctx).Save(w).Error
	if err != nil {
		err = erax.Wrap(err, "failed to update withdrawal")
		return erax.WithMeta(err, "layer", "DB")
	}

	return nil
}
