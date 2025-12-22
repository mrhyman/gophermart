package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/model"
)

//go:generate mockgen -source=balance.go -destination=mocks/mock_balance_repository.go -package=mocks

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (int, int, error)
	Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum int) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error)
}

func (r *Repository) GetUserBalance(ctx context.Context, userID uuid.UUID) (int, int, error) {
	query := `
		SELECT u.balance, COALESCE(w.withdrawn, 0) AS withdrawn
		FROM users u 
		LEFT JOIN (
			SELECT user_id, SUM(sum) AS withdrawn 
			FROM withdraws 
			GROUP BY user_id
		) AS w ON u.id = w.user_id 
		WHERE u.id = $1
	`

	var current, withdrawn int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&current, &withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, model.ErrNotFound
		}
		return 0, 0, err
	}

	return current, withdrawn, nil
}

func (r *Repository) Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance int
	query := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}

	if balance < sum {
		return model.ErrInsufficientFunds
	}

	updateQuery := `UPDATE users SET balance = balance - $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, sum, userID)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO withdraws (user_id, order_id, sum, processed_at) VALUES ($1, $2, $3, NOW())`
	_, err = tx.ExecContext(ctx, insertQuery, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error) {
	query := `
		SELECT id, user_id, order_id, sum, processed_at 
		FROM withdraws 
		WHERE user_id = $1 
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		err := rows.Scan(&w.ID, &w.UserID, &w.OrderID, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
