package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrhyman/gophermart/internal/client"
	"github.com/mrhyman/gophermart/internal/config"
	"github.com/mrhyman/gophermart/internal/handler"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository"
	"github.com/mrhyman/gophermart/internal/server"
	"github.com/mrhyman/gophermart/internal/service"
	"github.com/mrhyman/gophermart/internal/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	log := logger.New()
	ctx = logger.WithinContext(ctx, log)

	defer log.Sync()

	cfg := config.Load(ctx)

	repo := initRepo(ctx, cfg)
	defer repo.Close()

	svc := service.New(repo)
	h := handler.New(*svc, cfg.HashKey)
	s := server.New(cfg, *h)

	accrualClient := client.NewAccrualClient(cfg.AccrualAddress)
	w := worker.NewAccrualWorker(
		repo,
		repo,
		accrualClient,
		config.WorkerPollInterval,
		config.WorkerBatchSize,
		config.WorkerPoolSize,
	)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return w.Start(ctx)
	})

	g.Go(func() error {
		return s.Start(ctx)
	})

	log.Info("Application started")

	if err := g.Wait(); err != nil {
		log.With("err", err.Error()).Error(model.ErrServerCrash)
	}
	log.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.With("err", err.Error()).Error(model.ErrServerCrash)
	}

	log.Info("Application stopped")
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
