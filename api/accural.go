package api

import "github.com/mrhyman/gophermart/internal/model"

type AccrualResponse struct {
	Order   string              `json:"order"`
	Status  model.AccrualStatus `json:"status"`
	Accrual *int                `json:"accrual,omitempty"`
}
