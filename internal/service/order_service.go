package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository"
	"github.com/mrhyman/gophermart/internal/util"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, number string) (*model.Order, error) {
	if !util.ValidateLuhn(number) {
		return nil, model.ErrInvalidOrderNumber
	}

	existingOrder, err := s.repo.GetByNumber(ctx, number)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return existingOrder, model.ErrOrderAlreadyUploaded
		}

		return nil, model.ErrOrderUploadedByAnotherUser
	}

	return s.repo.Create(ctx, userID, number)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	return s.repo.GetUserOrders(ctx, userID)
}
