package handler

import (
	"github.com/mrhyman/gophermart/internal/service"
)

type HTTPHandler struct {
	svc service.Service
}

func New(
	svc service.Service,

) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}
