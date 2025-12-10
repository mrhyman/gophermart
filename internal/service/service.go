package service

import "github.com/mrhyman/gophermart/internal/repository"

type Service struct {
	repo repository.Repository
}

func New(
	repo repository.Repository,

) *Service {
	return &Service{
		repo: repo,
	}
}
