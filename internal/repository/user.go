package repository

import (
	"context"

	"github.com/mrhyman/gophermart/internal/model"
)

func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	// логика создания пользователя
	return nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	// логика получения пользователя
	return nil, nil
}
