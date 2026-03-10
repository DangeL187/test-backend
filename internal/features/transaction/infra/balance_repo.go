package infra

import (
	"context"

	"github.com/DangeL187/erax"
	"gorm.io/gorm"

	"back/internal/features/transaction/domain"
	"back/internal/infra/database"
)

type BalanceRepo struct {
	db *gorm.DB
}

func NewBalanceRepo(db *gorm.DB) *BalanceRepo {
	return &BalanceRepo{db: db}
}

func (r *BalanceRepo) GetBalance(ctx context.Context, userID int, currency string) (float64, error) {
	db := database.GetDB(ctx, r.db)

	var balance float64

	err := db.WithContext(ctx).
		Model(&domain.Balance{}).
		Where("user_id = ? AND currency = ?", userID, currency).
		Select("balance").
		Scan(&balance).Error
	if err != nil {
		err = erax.Wrap(err, "failed to get balance")
		return 0, erax.WithMeta(err, "layer", "DB")
	}

	return balance, nil
}

func (r *BalanceRepo) UpdateBalance(ctx context.Context, userID int, currency string, amount float64) (float64, error) {
	db := database.GetDB(ctx, r.db)

	var newBalance float64

	if amount < 0 {
		// Concurrent balance update (-):
		res := db.WithContext(ctx).
			Raw(`
	            UPDATE balances
	            SET balance = balance + $1
	            WHERE user_id = $2 AND currency = $3
	            AND balance + $1 >= 0
	            RETURNING balance
        	`, amount, userID, currency).
			Scan(&newBalance)
		if res.Error != nil {
			err := erax.Wrap(res.Error, "failed to withdraw balance")
			return 0, erax.WithMeta(err, "layer", "DB")
		} else if res.RowsAffected == 0 {
			return 0, domain.ErrInsufficientBalance // Not enough balance or balance does not exist
		}

		return newBalance, nil
	}

	// Concurrent balance update (+); create balance if it doesn't exist:
	err := db.WithContext(ctx).
		Raw(`
	        WITH updated AS (
				UPDATE balances
				SET balance = balance + $1
				WHERE user_id = $2 AND currency = $3
				RETURNING balance
			)
			INSERT INTO balances(user_id, balance, currency)
			SELECT $2, $1, $3
			WHERE NOT EXISTS (SELECT 1 FROM updated)
			RETURNING balance
		`, amount, userID, currency).
		Scan(&newBalance).Error
	if err != nil {
		err = erax.Wrap(err, "failed to deposit balance")
		return 0, erax.WithMeta(err, "layer", "DB")
	}

	return newBalance, nil
}
