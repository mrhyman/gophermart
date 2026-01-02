package service

import "github.com/mrhyman/gophermart/internal/repository"

type Service struct {
	User    *UserService
	Order   *OrderService
	Balance *BalanceService
}

func New(repos *repository.Repos) *Service {
	return &Service{
		User:    NewUserService(repos.User),
		Order:   NewOrderService(repos.Order),
		Balance: NewBalanceService(repos.Balance),
	}
}
