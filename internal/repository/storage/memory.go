package storage

import (
	"sync"

	"github.com/mrhyman/gophermart/internal/model"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	order map[string]model.Order
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		order: make(map[string]model.Order),
	}
}

func (ms *MemoryStorage) Ping() error {
	return nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}
