package service

import (
	"context"

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

func (s *UserService) Register(ctx context.Context, login, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	return s.repo.CreateUser(ctx, login, string(hash))
}

func (s *UserService) Login(ctx context.Context, login, password string) error {
	dbUser, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}

	if err := auth.CheckPassword(password, dbUser.Password); err != nil {
		return model.ErrInvalidCredentials
	}

	return nil
}
