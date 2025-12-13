package handler

import (
	"github.com/mrhyman/gophermart/internal/service"
)

type HTTPHandler struct {
	User *UserHandler
}

func New(svc service.Service) *HTTPHandler {
	return &HTTPHandler{
		User: NewUserHandler(&svc),
	}
}
