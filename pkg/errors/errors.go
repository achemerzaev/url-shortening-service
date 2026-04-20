package errors

import "errors"

var (
	ErrEmailExists        = errors.New("EMAIL_EXISTS email already exists")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS incorrect email or password")
	ErrInvalidToken       = errors.New("INVALID_TOKEN authentification failed, please log in or refresh token")
	ErrForbidden          = errors.New("FORBIDDEN user dont own this resource")
	ErrNotFound           = errors.New("NOT_FOUND user/url not found")
	ErrGeneratingJWT      = errors.New("JWT_ERROR authorization error on server side")
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func New(code string, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code string, message string, err error) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{Code: code, Message: message, Err: err}
}
