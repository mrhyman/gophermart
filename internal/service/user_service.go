package service

import (
	"context"

	"github.com/mrhyman/gophermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repository.Repository
}

func NewUserService(repo repository.Repository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.CreateUser(ctx, login, string(hash))
}
