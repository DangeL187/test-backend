package config

import (
	"fmt"
	"os"
	"time"

	"github.com/DangeL187/erax"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnectTimeout time.Duration
	PostgresDSN      string
	Token            string
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		zap.L().Info("[env] no .env file found, skipping")
	}

	postgresDSN, token, err := loadVars()
	if err != nil {
		return nil, erax.Wrap(err, "failed to load environment variables")
	}

	return &Config{
		DBConnectTimeout: 1 * time.Minute,
		PostgresDSN:      postgresDSN,
		Token:            token,
	}, nil
}

func loadVars() (string, string, error) {
	vars := map[string]string{
		"JWT_BEARER_TOKEN":  os.Getenv("JWT_BEARER_TOKEN"),
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_SSL_MODE": os.Getenv("POSTGRES_SSL_MODE"),
	}

	for name, value := range vars {
		if value == "" {
			return "", "", fmt.Errorf("missing required env var: %s", name)
		}
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		vars["POSTGRES_HOST"],
		vars["POSTGRES_PORT"],
		vars["POSTGRES_USER"],
		vars["POSTGRES_PASSWORD"],
		vars["POSTGRES_DB"],
		vars["POSTGRES_SSL_MODE"],
	)

	return dsn, vars["JWT_BEARER_TOKEN"], nil
}
