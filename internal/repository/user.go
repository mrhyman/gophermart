package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/model"
)

//go:generate mockgen -source=user.go -destination=mocks/mock_user_repository.go -package=mocks

type UserRepository interface {
	CreateUser(ctx context.Context, login string, password string) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

func (r *Repository) CreateUser(ctx context.Context, login string, password string) error {
	user := &model.User{
		ID:       uuid.New(),
		Login:    login,
		Password: password,
	}

	query := `INSERT INTO users (id, login, password) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, user.ID, user.Login, user.Password); err != nil {
		return r.convertPgError(ctx, "user", user.Login, err)
	}

	return nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT id, login, password FROM users WHERE login = $1`

	var user model.User
	if err := r.db.GetContext(ctx, &user, query, login); err != nil {
		return nil, r.convertPgError(ctx, "user", login, err)
	}

	return &user, nil
}
