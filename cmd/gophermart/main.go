package main

import (
	"context"
	"time"

	"github.com/mrhyman/gophermart/internal/client"
	"github.com/mrhyman/gophermart/internal/config"
	"github.com/mrhyman/gophermart/internal/handler"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/repository"
	"github.com/mrhyman/gophermart/internal/server"
	"github.com/mrhyman/gophermart/internal/service"
	"github.com/mrhyman/gophermart/internal/worker"
)

func main() {
	ctx := context.Background()
	log := logger.New()
	logger.WithinContext(ctx, log)

	defer log.Sync()

	cfg := config.Load(ctx)

	repo := initRepo(ctx, cfg)
	defer repo.Close()

	svc := service.New(repo)
	h := handler.New(*svc, cfg.HashKey)
	s := server.New(cfg, *h)

	accrualClient := client.NewAccrualClient(cfg.AccrualAddress)
	worker := worker.NewAccrualWorker(
		repo,
		repo,
		accrualClient,
		5*time.Second,
		10,
		3,
	)

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go worker.Start(workerCtx)

	s.Start(ctx)
}

func initRepo(ctx context.Context, cfg config.AppConfig) *repository.Repository {
	log := logger.FromContext(ctx)

	repo, err := repository.NewRepository(cfg.DBURI)

	if err != nil {
		log.With("err", err.Error()).Fatal()
	}

	err = repo.MigrateUp("migrations", cfg.DBURI)
	if err != nil {
		log.With("err", err.Error()).Fatal()
	}
	log.Info("DB connection set. Migrations applied successfully")

	return repo
}
