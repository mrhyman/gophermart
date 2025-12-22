package model

import (
	"errors"
	"fmt"
)

var (
	ErrLoggerSetup                = errors.New("logger setup error")
	ErrCookieDecoding             = errors.New("can't decode cookie")
	ErrCookieEncoding             = errors.New("can't encode cookie")
	ErrCompressReading            = errors.New("compress reading error")
	ErrUnknownAccrualStatus       = errors.New("unknown accrual status")
	ErrUnknownUser                = errors.New("userID is not provided")
	ErrInvalidRequestParams       = errors.New("invalid request params")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrResponseDecode             = errors.New("can't decode response")
	ErrWentWrong                  = errors.New("something went wrong")
	ErrNotFound                   = errors.New("entity not found")
	ErrInvalidOrderNumber         = errors.New("invalid order number")
	ErrOrderAlreadyUploaded       = errors.New("order already uploaded")
	ErrOrderUploadedByAnotherUser = errors.New("order uploaded by another user")
	ErrResponseEncoding           = errors.New("can't encode response")
	ErrInsufficientFunds          = errors.New("insufficient funds")
	// accrual errors
	ErrAccrualRequestCreateFailed = errors.New("can't create accrual request")
	ErrAccrualRequestSendFailed   = errors.New("can't send accrual request")
	ErrOrderNotRegistered         = errors.New("order not registered")
	ErrAccrualTooManyRequests     = errors.New("too many accrual requests")
	ErrAccrualInternalError       = errors.New("accrual internal error")
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
