package main

import (
	"context"

	"github.com/mrhyman/gophermart/internal/config"
	"github.com/mrhyman/gophermart/internal/handler"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/repository"
	"github.com/mrhyman/gophermart/internal/server"
	"github.com/mrhyman/gophermart/internal/service"
)

func main() {
	ctx := context.Background()
	log := logger.New()
	logger.WithinContext(ctx, log)

	defer log.Sync()

	cfg := config.Load(ctx)

	repo := initStorage(ctx, cfg)
	defer repo.Close()

	svc := service.New(repo)
	h := handler.New(*svc)
	s := server.New(cfg, *h)

	s.Start(ctx)
}

func initStorage(ctx context.Context, cfg config.AppConfig) repository.Repository {
	log := logger.FromContext(ctx)

	repo, err := repository.NewRepository(cfg.StoragePath)

	if err != nil {
		log.With("err", err.Error()).Fatal()
	}

	err = repo.MigrateUp("migrations", cfg.DBDSN)
	if err != nil {
		log.With("err", err.Error()).Fatal()
	}
	log.Info("DB connection set. Migrations applied successfully")

	return *repo
}
