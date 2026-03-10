package app

import (
	"github.com/DangeL187/erax"

	trHandler "back/internal/features/transaction/handler"
	trInfra "back/internal/features/transaction/infra"
	trUC "back/internal/features/transaction/usecase"
	"back/internal/infra/database"
	"back/internal/shared/config"
)

type App struct {
	Config             *config.Config
	TransactionHandler *trHandler.TransactionHandler
}

func NewApp() (*App, error) {
	app := &App{}

	var err error
	app.Config, err = config.NewConfig()
	if err != nil {
		return nil, erax.Wrap(err, "failed to load config")
	}

	db, err := database.NewPostgres(app.Config)
	if err != nil {
		return nil, erax.Wrap(err, "failed to connect to DB")
	}

	balanceRepo := trInfra.NewBalanceRepo(db)
	ledgerRepo := trInfra.NewLedgerRepo(db)
	withdrawalRepo := trInfra.NewWithdrawalRepo(db)
	txManager := database.NewTxManager(db)

	transactionUseCase := trUC.NewTransactionUseCase(balanceRepo, ledgerRepo, withdrawalRepo, txManager)
	app.TransactionHandler = trHandler.NewTransactionHandler(transactionUseCase)

	return app, nil
}
