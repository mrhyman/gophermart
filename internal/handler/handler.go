package handler

import (
	"github.com/mrhyman/gophermart/internal/service"
)

type HTTPHandler struct {
	User   *UserHandler
	Order  *OrderHandler
	Secret string
}

func New(svc service.Service, secret string) *HTTPHandler {
	return &HTTPHandler{
		User:  NewUserHandler(&svc, secret),
		Order: NewOrderHandler(&svc),
	}
}
