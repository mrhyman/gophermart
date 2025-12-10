package repository

import (
	"context"

	"github.com/mrhyman/gophermart/internal/repository/storage"
)

type OrderRepository interface {
	Ping(ctx context.Context) error
}

type OrderRepo struct {
	storage storage.Storage
}

func NewURLRepository(s storage.Storage) *OrderRepo {
	return &OrderRepo{storage: s}
}

func (r *OrderRepo) Ping(ctx context.Context) error {
	return r.storage.Ping()
}
