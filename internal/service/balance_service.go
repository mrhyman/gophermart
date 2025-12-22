package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository"
)

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID uuid.UUID) (int, int, error) {
	return s.repo.GetUserBalance(ctx, userID)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum int) error {
	return s.repo.Withdraw(ctx, userID, orderNumber, sum)
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error) {
	return s.repo.GetWithdrawals(ctx, userID)
}
