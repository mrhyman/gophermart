package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/model"
)

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
	// логика получения пользователя
	return nil, nil
}
