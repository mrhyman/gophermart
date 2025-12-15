package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/gophermart/internal/model"
)

//go:generate mockgen -source=order.go -destination=mocks/mock_order_repository.go -package=mocks

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, number string) (*model.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
	CountOrdersByStatus(ctx context.Context, status model.OrderStatus) (int, error)
	GetOrdersForProcessing(ctx context.Context, tx *sqlx.Tx, status model.OrderStatus, limit int, offset int) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus, accrual int) error
	UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, orderID uuid.UUID, status model.OrderStatus, accrual int) error
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

func (r *Repository) CreateOrder(ctx context.Context, userID uuid.UUID, number string) (*model.Order, error) {
	order, err := model.NewOrder(uuid.New(), userID, number, model.OrderStatusNew, 0, time.Now())

	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO orders (id, user_id, number, status, created_at) 
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Number,
		order.Status,
	).Scan(&order.CreatedAt)

	if err != nil {
		return nil, r.convertPgError(ctx, "order", number, err)
	}

	return order, nil
}

func (r *Repository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `
		SELECT id, user_id, number, status, created_at 
		FROM orders 
		WHERE number = $1
	`

	var order model.Order
	err := r.db.GetContext(ctx, &order, query, number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (r *Repository) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	query := `
		SELECT id, user_id, number, status, created_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`

	var orders []*model.Order
	err := r.db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) CountOrdersByStatus(ctx context.Context, status model.OrderStatus) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE status = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, status)
	return count, err
}

func (r *Repository) GetOrdersForProcessing(
	ctx context.Context,
	tx *sqlx.Tx,
	status model.OrderStatus,
	limit int,
	offset int,
) ([]*model.Order, error) {
	query := `
		SELECT * 
		FROM orders 
		WHERE status = $1 
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
		FOR UPDATE SKIP LOCKED
	`

	var orders []*model.Order
	err := tx.SelectContext(ctx, &orders, query, status, limit, offset)

	return orders, err
}

func (r *Repository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
}

func (r *Repository) UpdateOrderStatusTx(
	ctx context.Context,
	tx *sqlx.Tx,
	orderID uuid.UUID,
	status model.OrderStatus,
	accrual int,
) error {
	query := `
		UPDATE orders 
		SET status = $1, accrual = $2 
		WHERE id = $3
	`

	_, err := tx.ExecContext(ctx, query, status, accrual, orderID)
	return err
}

func (r *Repository) UpdateOrderStatus(
	ctx context.Context,
	orderID uuid.UUID,
	status model.OrderStatus,
	accrual int,
) error {
	query := `
		UPDATE orders 
		SET status = $1, accrual = $2 
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, accrual, orderID)
	return err
}
