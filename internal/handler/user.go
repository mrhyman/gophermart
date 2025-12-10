package handler

import (
	"net/http"
)

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	// логика регистрации с использованием h.userService
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	// логика логина
}
