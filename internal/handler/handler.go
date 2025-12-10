package handler

import (
	"github.com/mrhyman/gophermart/internal/service"
)

type HTTPHandler struct {
	svc service.OrderService
}

func New(
	svc service.OrderService,

) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}
