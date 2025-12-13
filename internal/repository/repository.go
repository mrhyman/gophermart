package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mrhyman/gophermart/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Repository{db}, nil
}

func (ds *Repository) Ping() error {
	return ds.db.Ping()
}

func (ds *Repository) Close() error {
	return ds.db.Close()
}

func (ds *Repository) MigrateUp(migrationsDir, dsn string) error {
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

func (ds *Repository) convertPgError(ctx context.Context, entity, id string, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return model.NewAlreadyExistsError(entity, id, err)
	}
	return err
}
