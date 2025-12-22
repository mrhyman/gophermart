package worker

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/gophermart/internal/client"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository"
)

type AccrualWorker struct {
	orderRepo     repository.OrderRepository
	userRepo      repository.UserRepository
	accrualClient *client.AccrualClient
	pollInterval  time.Duration
	batchSize     int
	poolSize      int
}

func NewAccrualWorker(
	repo repository.OrderRepository,
	userRepo repository.UserRepository,
	accrualClient *client.AccrualClient,
	pollInterval time.Duration,
	batchSize int,
	poolSize int,
) *AccrualWorker {
	return &AccrualWorker{
		orderRepo:     repo,
		userRepo:      userRepo,
		accrualClient: accrualClient,
		pollInterval:  pollInterval,
		batchSize:     batchSize,
		poolSize:      poolSize,
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	for i := 0; i < w.poolSize; i++ {
		go w.worker(ctx, i, w.pollInterval, w.batchSize)
	}
}

func (w *AccrualWorker) worker(ctx context.Context, workerID int, pollInterval time.Duration, batchSize int) {
	log := logger.FromContext(ctx).With("worker_id", workerID)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping accrual worker")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx, batchSize); err != nil {
				log.With("err", err.Error()).Error()
			}
		}
	}
}

func (w *AccrualWorker) processBatch(ctx context.Context, batchSize int) error {
	log := logger.FromContext(ctx)

	tx, err := w.orderRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orders, err := w.orderRepo.GetOrdersForProcessing(ctx, tx, w.batchSize)

	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	for _, order := range orders {
		if err := w.processOrder(ctx, tx, order); err != nil {
			log.With("order", order.Number, "err", err.Error()).Error()

			if errors.Is(err, model.ErrAccrualTooManyRequests) {
				return err
			}
		}
	}

	return tx.Commit()
}

func (w *AccrualWorker) processOrder(ctx context.Context, tx *sqlx.Tx, order *model.Order) error {
	log := logger.FromContext(ctx)

	accrualResp, err := w.accrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotRegistered) {
			log.With("order", order.Number).Debug()
			return nil
		}
		return err
	}

	newStatus, err := model.MapAccrualStatusToOrderStatus(accrualResp.Status)
	if err != nil {
		return err
	}

	// if newStatus == order.Status && accrualResp.Accrual == nil {
	// 	return nil
	// }

	var accrual int
	if accrualResp.Accrual != nil {
		accrual = int(*accrualResp.Accrual * 100)
	}

	log.With("before_update", "accrual", accrual, "status", newStatus).Debug()

	if err := w.updateOrderAndBalance(ctx, tx, order, newStatus, accrual); err != nil {
		return err
	}

	log.With("order", order.Number, "status", newStatus, "accrual", accrual).Info()
	return nil
}

func (w *AccrualWorker) updateOrderAndBalance(
	ctx context.Context,
	tx *sqlx.Tx,
	order *model.Order,
	newStatus model.OrderStatus,
	accrual int,
) error {
	log := logger.FromContext(ctx)

	if err := w.orderRepo.UpdateOrderStatusTx(ctx, tx, order.ID, newStatus, accrual); err != nil {
		return err
	}

	log.With("inside_update", "accrual", accrual, "status", newStatus).Debug()

	return w.userRepo.AddBalanceTx(ctx, tx, order.UserID, accrual)
}
