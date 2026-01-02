package repository

import (
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type Repos struct {
	db      *sqlx.DB
	User    *UserRepo
	Order   *OrderRepo
	Balance *BalanceRepo
}

func NewRepos(dsn string) (*Repos, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Repos{
		db:      db,
		User:    NewUserRepository(db),
		Order:   NewOrderRepository(db),
		Balance: NewBalanceRepository(db),
	}, nil
}

func (r *Repos) Ping() error {
	return r.db.Ping()
}

func (r *Repos) Close() error {
	return r.db.Close()
}

func (r *Repos) MigrateUp(migrationsDir, dsn string) error {
	driver, err := postgres.WithInstance(r.db.DB, &postgres.Config{})
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
