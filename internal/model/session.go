package model

import (
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

func NewSession(
	userID ulid.ULID,
	refreshToken string,
	refreshTokenExpiredAt time.Time,
) Session {
	return Session{
		ID:                    NewID(),
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiredAt: refreshTokenExpiredAt,
		CreatedAt:             time.Now(),
	}
}
