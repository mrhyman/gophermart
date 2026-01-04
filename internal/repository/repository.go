package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mrhyman/gophermart/internal/model"
)

type TxFunc func(tx *sqlx.Tx) error

type Entity interface {
	TableName() string
}

type GenericRepository[T Entity] struct {
	db *sqlx.DB
}

func NewGenericRepository[T Entity](db *sqlx.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

func (r *GenericRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, entity.TableName())

	err := r.db.GetContext(ctx, &entity, query, id)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entity T
	var entities []T
	query := fmt.Sprintf(`SELECT * FROM %s`, entity.TableName())

	err := r.db.SelectContext(ctx, &entities, query)
	return entities, err
}

func (r *GenericRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, entity.TableName())

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *GenericRepository[T]) Count(ctx context.Context) (int, error) {
	var entity T
	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, entity.TableName())

	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *GenericRepository[T]) WithTx(ctx context.Context, fn TxFunc) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *GenericRepository[T]) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
}

func (r *GenericRepository[T]) DB() *sqlx.DB {
	return r.db
}

func (r *GenericRepository[T]) convertPgError(ctx context.Context, entity, id string, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return model.NewAlreadyExistsError(entity, id, err)
	}
	return err
}
