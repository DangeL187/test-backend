package infra

import (
	"context"

	"github.com/DangeL187/erax"
	"gorm.io/gorm"

	"back/internal/features/transaction/domain"
)

type LedgerRepo struct {
	db *gorm.DB
}

func NewLedgerRepo(db *gorm.DB) *LedgerRepo {
	return &LedgerRepo{db: db}
}

func (r *LedgerRepo) Create(ctx context.Context, entry *domain.LedgerEntry) error {
	err := r.db.WithContext(ctx).Create(entry).Error
	if err != nil {
		err = erax.Wrap(err, "failed to create ledger entry")
		return erax.WithMeta(err, "layer", "DB")
	}

	return nil
}
