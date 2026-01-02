package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
	Balance  int       `db:"balance"`
}

func NewUser(
	ID uuid.UUID,
	login string,
	password string,
) (*User, error) {
	user := &User{
		ID:       ID,
		Login:    login,
		Password: password,
	}

	return user, nil
}

func (User) TableName() string { return "users" }
