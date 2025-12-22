package config

import (
	"context"
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/mrhyman/gophermart/internal/logger"
)

const (
	DefaultRunAddress     = "localhost:8080"
	DefaultDBURI          = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	DefaultAccrualAddress = "localhost:9090"
	DefaultHashKey        = "qwerty12345"
	WorkerPollInterval    = 1 * time.Second
	WorkerBatchSize       = 10
	WorkerPoolSize        = 3
	AccuralRequestTimeout = 10 * time.Second
)

type AppConfig struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DBURI          string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	HashKey        string `env:"HASH_KEY"`
}

func Load(ctx context.Context) AppConfig {
	var cfg AppConfig

	log := logger.FromContext(ctx)

	runFlag := flag.String("a", DefaultRunAddress, "HTTP server address, e.g. localhost:8888")
	dbFlag := flag.String("d", DefaultDBURI, "Database connection string. postgres://postgres:postgres@localhost:5432/postgres")
	accFlag := flag.String("r", DefaultAccrualAddress, "Accrual service address, e.g. localhost:8080")
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

	if cfg.AccrualAddress == "" {
		cfg.AccrualAddress = *accFlag
	}

	if cfg.HashKey == "" {
		cfg.HashKey = *hashKey
	}

	return cfg
}
