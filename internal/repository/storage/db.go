package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mrhyman/gophermart/internal/model"
)

type DBStorage struct {
	db *sqlx.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DBStorage{db}, nil
}

func (ds *DBStorage) Ping() error {
	return ds.db.Ping()
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}

func (ds *DBStorage) MigrateUp(migrationsDir, dsn string) error {
	driver, err := postgres.WithInstance(ds.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (ds *DBStorage) convertPgError(ctx context.Context, l model.Order, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.Constraint == "idx_links_original_url" {
			// existing, getErr := ds.GetByOriginalURL(ctx, l.OriginalURL)
			// if getErr == nil {
			// 	return model.NewAlreadyExistsError(existing.ShortURL, err)
			// }
			return nil
		}
	}

	return err
}
