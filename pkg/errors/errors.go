package errors

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("resource already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
	ErrValidation         = errors.New("validation error")
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
