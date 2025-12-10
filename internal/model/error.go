package model

import (
	"errors"
	"fmt"
)

var (
	ErrLoggerSetup     = errors.New("logger setup error")
	ErrCookieDecoding  = errors.New("can't decode cookie")
	ErrCookieEncoding  = errors.New("can't encode cookie")
	ErrCompressReading = errors.New("compress reading error")
	ErrUnknownUser     = errors.New("userID is not provided")
	ErrWentWrong       = errors.New("something went wrong")
)

type AlreadyExistsError struct {
	OrderID string
	Err     error
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("order with such id already exists: %s", e.OrderID)
}

func (e *AlreadyExistsError) Unwrap() error {
	return e.Err
}

func NewAlreadyExistsError(orderID string, err error) error {
	return &AlreadyExistsError{
		OrderID: orderID,
		Err:     err,
	}
}
