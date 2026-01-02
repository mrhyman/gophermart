package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/gophermart/internal/model"
)

//go:generate mockgen -source=user.go -destination=mocks/mock_user_repository.go -package=mocks

type UserRepository interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	Create(ctx context.Context, user model.User) error
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	AddBalance(ctx context.Context, userID uuid.UUID, amount int) error
	AddBalanceTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount int) error
	GetBalance(ctx context.Context, userID uuid.UUID) (int, error)
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type UserRepo struct {
	*GenericRepository[model.User]
}

func NewUserRepository(db *sqlx.DB) *UserRepo {
	return &UserRepo{
		GenericRepository: NewGenericRepository[model.User](db),
	}
}

func (r *UserRepo) Create(ctx context.Context, user model.User) error {
	query := `INSERT INTO users (id, login, password) VALUES ($1, $2, $3)`
	if _, err := r.DB().ExecContext(ctx, query, user.ID, user.Login, user.Password); err != nil {
		return r.convertPgError(ctx, "user", user.Login, err)
	}

	return nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT id, login, password FROM users WHERE login = $1`

	var user model.User
	if err := r.db.GetContext(ctx, &user, query, login); err != nil {
		return nil, r.convertPgError(ctx, "user", login, err)
	}

	return &user, nil
}

func (r *UserRepo) AddBalance(ctx context.Context, userID uuid.UUID, amount int) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, amount, userID)
	return err
}

func (r *UserRepo) AddBalanceTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount int) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, amount, userID)
	return err
}

func (r *UserRepo) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT balance FROM users WHERE id = $1`

	var balance int
	err := r.db.GetContext(ctx, &balance, query, userID)
	if err != nil {
		return 0, err
	}

	return balance, nil
}
