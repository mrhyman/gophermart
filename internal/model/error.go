package model

import (
	"errors"
	"fmt"
)

var (
	ErrLoggerSetup          = errors.New("logger setup error")
	ErrCookieDecoding       = errors.New("can't decode cookie")
	ErrCookieEncoding       = errors.New("can't encode cookie")
	ErrCompressReading      = errors.New("compress reading error")
	ErrUnknownUser          = errors.New("userID is not provided")
	ErrInvalidRequestParams = errors.New("invalid request params")
	ErrWentWrong            = errors.New("something went wrong")
)

type AlreadyExistsError struct {
	Entity string
	ID     string
	Err    error
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s with id '%s' already exists", e.Entity, e.ID)
}

func (e *AlreadyExistsError) Unwrap() error {
	return e.Err
}

func NewAlreadyExistsError(entity, id string, err error) error {
	return &AlreadyExistsError{
		Entity: entity,
		ID:     id,
		Err:    err,
	}
}
