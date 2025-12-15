package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/service"
)

type OrderHandler struct {
	os *service.OrderService
}

func NewOrderHandler(svc *service.Service) *OrderHandler {
	return &OrderHandler{
		os: svc.Order,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		log.With("err", "invalid content type").Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		log.With("err", "empty order number").Warn()
		http.Error(w, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
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

	_, err = h.os.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidOrderNumber):
			log.With("err", err.Error()).Warn()
			w.WriteHeader(http.StatusUnprocessableEntity)
			return

		case errors.Is(err, model.ErrOrderAlreadyUploaded):
			log.With("order", orderNumber).Info("order already uploaded by this user")
			w.WriteHeader(http.StatusOK)
			return

		case errors.Is(err, model.ErrOrderUploadedByAnotherUser):
			log.With("err", err.Error()).Warn()
			w.WriteHeader(http.StatusConflict)
			return

		default:
			log.With("err", err.Error()).Error()
			http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
