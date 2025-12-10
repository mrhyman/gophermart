package service

import "github.com/mrhyman/gophermart/internal/repository"

type OrderService struct {
	base string
	repo repository.OrderRepository
}

func NewOrderService(base string, repo repository.OrderRepository) *OrderService {
	return &OrderService{base: base, repo: repo}
}
