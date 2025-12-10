package service

import "github.com/mrhyman/gophermart/internal/repository"

type UserService struct {
	repo repository.Repository
}

func NewOrderService(repo repository.Repository) *UserService {
	return &UserService{repo: repo}
}
