package model

import (
	"time"

	"github.com/google/uuid"
)

type AccrualStatus string

const (
	AccrualStatusNew        AccrualStatus = "NEW"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

func MapAccrualStatusToOrderStatus(accrualStatus AccrualStatus) (OrderStatus, error) {
	switch accrualStatus {
	case AccrualStatusNew:
		return OrderStatusNew, nil
	case AccrualStatusProcessing:
		return OrderStatusProcessing, nil
	case AccrualStatusInvalid:
		return OrderStatusInvalid, nil
	case AccrualStatusProcessed:
		return OrderStatusProcessed, nil
	default:
		return "", ErrUnknownAccrualStatus
	}
}

type Order struct {
	ID        uuid.UUID   `db:"id" json:"id"`
	UserID    uuid.UUID   `db:"user_id" json:"user_id"`
	Number    string      `db:"number" json:"number"`
	Status    OrderStatus `db:"status" json:"status"`
	Accrual   int         `db:"accrual" json:"accrual"`
	CreatedAt time.Time   `db:"created_at" json:"uploaded_at"`
}

func NewOrder(
	ID uuid.UUID,
	userID uuid.UUID,
	number string,
	status OrderStatus,
	accrual int,
	createdAt time.Time,
) (*Order, error) {
	user := &Order{
		ID:        ID,
		UserID:    userID,
		Number:    number,
		Status:    status,
		Accrual:   accrual,
		CreatedAt: createdAt,
	}

	return user, nil
}
