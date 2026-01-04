package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/api"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/service"
	"github.com/mrhyman/gophermart/internal/util"
)

type BalanceHandler struct {
	bs *service.BalanceService
}

func NewBalanceHandler(svc *service.Service) *BalanceHandler {
	return &BalanceHandler{
		bs: svc.Balance,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	userIDStr, ok := r.Context().Value(model.UserIDKey).(string)
	if !ok || userIDStr == "" {
		log.With("err", model.ErrUnknownUser).Warn()
		http.Error(w, model.ErrUnknownUser.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	current, withdrawn, err := h.bs.GetUserBalance(r.Context(), userID)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	resp := api.UserBalanceResponse{
		Current:   util.RoundToTwoDecimals(float64(current) / 100),
		Withdrawn: util.RoundToTwoDecimals(float64(withdrawn) / 100),
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		log.With("err", err.Error())
		http.Error(w, model.ErrResponseEncoding.Error(), http.StatusBadRequest)
		return
	}
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	var req api.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	userIDStr, ok := r.Context().Value(model.UserIDKey).(string)
	if !ok || userIDStr == "" {
		log.With("err", model.ErrUnknownUser).Warn()
		http.Error(w, model.ErrUnknownUser.Error(), http.StatusUnauthorized)
		return
	}

	if !util.ValidateLuhn(req.Order) {
		log.With("err", model.ErrInvalidOrderNumber).Warn()
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	err = h.bs.Withdraw(r.Context(), userID, req.Order, int(req.Sum*100))
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	userIDStr, ok := r.Context().Value(model.UserIDKey).(string)
	if !ok || userIDStr == "" {
		log.With("err", model.ErrUnknownUser).Warn()
		http.Error(w, model.ErrUnknownUser.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.bs.GetWithdrawals(r.Context(), userID)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var resp = make([]api.WithdrawalListResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		resp = append(resp, api.WithdrawalListResponse{
			Order:       withdrawal.OrderID,
			Sum:         util.RoundToTwoDecimals(float64(withdrawal.Sum) / 100),
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		log.With("err", err.Error())
		http.Error(w, model.ErrResponseEncoding.Error(), http.StatusBadRequest)
		return
	}
}
