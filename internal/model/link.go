package model

import (
	"github.com/google/uuid"
)

type Order struct {
	ID     uuid.UUID `db:"id"`
	Status string    `db:"status"`
}

func NewLink(
	ID uuid.UUID,
	status string,
) (*Order, error) {
	link := &Order{
		ID:     ID,
		Status: status,
	}

	return link, nil
}
