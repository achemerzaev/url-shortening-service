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
