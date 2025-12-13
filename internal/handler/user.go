package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mrhyman/gophermart/api"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/service"
)

type UserHandler struct {
	us *service.UserService
}

func NewUserHandler(svc *service.Service) *UserHandler {
	return &UserHandler{
		us: svc.User,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest

	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	err := h.us.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		var existsErr *model.AlreadyExistsError

		switch {
		case errors.As(err, &existsErr):
			w.WriteHeader(http.StatusConflict)
			return

		default:
			log.With("err", err.Error()).Error()
			http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest

	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	err := h.us.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidCredentials):
			log.With("err", err.Error()).Warn()
			http.Error(w, err.Error(), http.StatusUnauthorized)

		default:
			log.With("err", err.Error()).Error()
			http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
