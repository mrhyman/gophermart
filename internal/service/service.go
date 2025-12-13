package service

import "github.com/mrhyman/gophermart/internal/repository"

type Service struct {
	User *UserService
	// Order   *OrderService
	// Balance *BalanceService
}

// type Service struct {
// 	repo repository.Repository
// }

func New(repo *repository.Repository) *Service {
	return &Service{
		User: NewUserService(*repo),
		// Order:   NewOrderService(repo),
		// Balance: NewBalanceService(repo),
	}
}
