package worker

import (
	"context"
	"errors"
	"sync"
	"time"

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
		accrualClient: accrualClient,
		pollInterval:  pollInterval,
		batchSize:     batchSize,
		poolSize:      poolSize,
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	log := logger.FromContext(ctx)
	log.With("pool_size", w.poolSize, "batch_size", w.batchSize).Info("starting accrual worker")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping accrual worker")
			return
		case <-ticker.C:
			if err := w.processBatches(ctx); err != nil {
				log.With("err", err.Error()).Error()
			}
		}
	}
}

func (w *AccrualWorker) processBatches(ctx context.Context) error {
	log := logger.FromContext(ctx)

	totalOrders, err := w.orderRepo.CountOrdersByStatus(ctx, model.OrderStatusNew)
	if err != nil {
		return err
	}

	if totalOrders == 0 {
		log.Debug("no orders to process")
		return nil
	}

	log.With("total_orders", totalOrders).Info("starting batch processing")

	jobs := make(chan int, w.poolSize)

	var wg sync.WaitGroup

	for i := 0; i < w.poolSize; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.worker(ctx, workerID, jobs)
		}(i)
	}

	go func() {
		for offset := 0; offset < totalOrders; offset += w.batchSize {
			select {
			case <-ctx.Done():
				return
			case jobs <- offset:
			}
		}
		close(jobs)
	}()

	wg.Wait()

	log.Info("batch processing completed")
	return nil
}

func (w *AccrualWorker) worker(ctx context.Context, workerID int, jobs <-chan int) {
	log := logger.FromContext(ctx).With("worker_id", workerID)

	for offset := range jobs {
		if err := w.processBatch(ctx, offset); err != nil {
			log.With("offset", offset, "err", err.Error()).Error()

			if errors.Is(err, model.ErrAccrualTooManyRequests) {
				log.Warn("too many requests, stopping worker")
				return
			}
		}
	}
}

func (w *AccrualWorker) processBatch(ctx context.Context, offset int) error {
	log := logger.FromContext(ctx)

	tx, err := w.orderRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orders, err := w.orderRepo.GetOrdersForProcessing(ctx, tx, model.OrderStatusNew, w.batchSize, offset)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	log.With("offset", offset, "count", len(orders)).Debug("processing batch")

	for _, order := range orders {
		if err := w.processOrder(ctx, order); err != nil {
			log.With("order", order.Number, "err", err.Error()).Error()

			if errors.Is(err, model.ErrAccrualTooManyRequests) {
				return err
			}
		}
	}

	return tx.Commit()
}

func (w *AccrualWorker) processOrder(ctx context.Context, order *model.Order) error {
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

	if newStatus == order.Status && accrualResp.Accrual == nil {
		return nil
	}

	var accrual int
	if accrualResp.Accrual != nil {
		accrual = *accrualResp.Accrual * 100
	}

	if err := w.updateOrderAndBalance(ctx, order, newStatus, accrual); err != nil {
		return err
	}

	log.With("order", order.Number, "status", newStatus, "accrual", accrual).Info()
	return nil
}

func (w *AccrualWorker) updateOrderAndBalance(
	ctx context.Context,
	order *model.Order,
	newStatus model.OrderStatus,
	accrual int,
) error {
	tx, err := w.orderRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := w.orderRepo.UpdateOrderStatusTx(ctx, tx, order.ID, newStatus, accrual); err != nil {
		return err
	}

	if newStatus == model.OrderStatusProcessed && accrual > 0 {
		if err := w.userRepo.AddBalanceTx(ctx, tx, order.UserID, accrual); err != nil {
			return err
		}
	}

	return tx.Commit()
}
