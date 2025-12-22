package model

import (
	"time"

	"github.com/google/uuid"
)

type Withdrawal struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	OrderID     string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func NewWithdrawal(
	ID uuid.UUID,
	userID uuid.UUID,
	orderID string,
	sum int,
	processedAt time.Time,
) (*Withdrawal, error) {
	w := &Withdrawal{
		ID:          ID,
		UserID:      userID,
		OrderID:     orderID,
		Sum:         sum,
		ProcessedAt: processedAt,
	}

	return w, nil
}
