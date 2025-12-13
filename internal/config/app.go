package config

import (
	"context"
	"flag"

	"github.com/caarlos0/env/v11"
	"github.com/mrhyman/gophermart/internal/logger"
)

const (
	DefaultRunAddress     = "localhost:9090"
	DefaultDBURI          = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	DefaultAccuralAddress = "localhost:8080"
	DefaultHashKey        = "qwerty12345"
)

type AppConfig struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DBURI          string `env:"DATABASE_URI"`
	AccuralAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	HashKey        string `env:"HASH_KEY"`
}

func Load(ctx context.Context) AppConfig {
	var cfg AppConfig

	log := logger.FromContext(ctx)

	runFlag := flag.String("a", DefaultRunAddress, "HTTP server address, e.g. localhost:8888")
	dbFlag := flag.String("d", DefaultDBURI, "Database connection string. postgres://postgres:postgres@localhost:5432/postgres")
	accFlag := flag.String("r", DefaultAccuralAddress, "Accural service address, e.g. localhost:8080")
	hashKey := flag.String("hk", DefaultHashKey, "Auth hash key. e.g. qwerty12345")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.With("err", err.Error()).Fatal()
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = *runFlag
	}

	if cfg.DBURI == "" {
		cfg.DBURI = *dbFlag
	}

	if cfg.AccuralAddress == "" {
		cfg.AccuralAddress = *accFlag
	}

	if cfg.HashKey == "" {
		cfg.HashKey = *hashKey
	}

	return cfg
}
