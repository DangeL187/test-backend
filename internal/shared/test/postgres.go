package test

import (
	"back/internal/shared/config"
	"context"
	"time"

	"github.com/DangeL187/erax"
	"gorm.io/gorm"

	"back/internal/features/transaction/domain"
	"back/internal/features/transaction/infra"
	"back/internal/infra/database"
)

type txManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type DB struct {
	DB             *gorm.DB
	BalanceRepo    domain.BalanceRepo
	LedgerRepo     domain.LedgerRepo
	WithdrawalRepo domain.WithdrawalRepo
	TxManager      txManager
	Teardown       func()
}

func SetupTestDB() (*DB, error) {
	cfg := &config.Config{
		DBConnectTimeout: 1 * time.Minute,
		PostgresDSN:      "host=localhost port=5432 user=myuser password=mypassword dbname=mydb sslmode=disable",
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		return nil, erax.Wrap(err, "failed to connect to database")
	}

	br := infra.NewBalanceRepo(db)
	lr := infra.NewLedgerRepo(db)
	wr := infra.NewWithdrawalRepo(db)
	txm := database.NewTxManager(db)

	return &DB{
		DB:             db,
		BalanceRepo:    br,
		LedgerRepo:     lr,
		WithdrawalRepo: wr,
		TxManager:      txm,
	}, nil
}
