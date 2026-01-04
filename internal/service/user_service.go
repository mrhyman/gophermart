package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/auth"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) (string, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &model.User{
		ID:       uuid.New(),
		Login:    login,
		Password: hash,
	}

	err = s.repo.Create(ctx, *user)
	if err != nil {
		return "", err
	}

	return user.ID.String(), nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (string, error) {
	dbUser, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if err := auth.CheckPassword(password, dbUser.Password); err != nil {
		return "", model.ErrInvalidCredentials
	}

	return dbUser.ID.String(), nil
}
