package model

import (
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

type Session struct {
	ID                    ulid.ULID
	UserID                ulid.ULID
	RefreshToken          string
	RefreshTokenExpiredAt time.Time
	CreatedAt             time.Time
}

var ErrInvalidArgument = errors.New("invalid arguments")

func NewSession(
	userID ulid.ULID,
	refreshToken string,
	refreshTokenExpiredAt time.Time,
) (sess Session, err error) {
	if refreshToken == "" || refreshTokenExpiredAt.IsZero() {
		return sess, ErrInvalidArgument
	}

	return Session{
		ID:                    NewID(),
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiredAt: refreshTokenExpiredAt,
		CreatedAt:             time.Now(),
	}, nil
}
