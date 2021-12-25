package usecase

import "errors"

// errors ...
var (
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInvalidToken        = errors.New("invalid token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)
