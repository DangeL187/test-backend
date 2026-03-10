package database

import (
	"context"
	"database/sql"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"

	"github.com/DangeL187/erax"

	"back/internal/shared/config"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
	defer cancel()

	var db *gorm.DB
	var err error

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, erax.Wrap(err, "failed to connect to database within timeout")
		default:
			db, err = gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				zap.L().Info("[DB] Postgres not ready yet. Retrying in 2s...")
				<-ticker.C
			}

			var sqlDB *sql.DB
			sqlDB, err = db.DB()
			if err != nil {
				return nil, err
			}

			sqlDB.SetMaxOpenConns(50)
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetConnMaxLifetime(time.Minute * 5)

			return db, nil
		}
	}
}
