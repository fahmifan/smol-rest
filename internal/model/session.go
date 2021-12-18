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

func (s *Session) SetRefreshToken(token string, expiredAt time.Time) error {
	if token == "" || expiredAt.IsZero() {
		return ErrInvalidArgument
	}

	s.RefreshToken = token
	s.RefreshTokenExpiredAt = expiredAt
	return nil
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.RefreshTokenExpiredAt)
}

var ErrInvalidArgument = errors.New("invalid arguments")

func NewSession(
	userID ulid.ULID,
) (sess Session, err error) {
	return Session{
		ID:        NewID(),
		UserID:    userID,
		CreatedAt: time.Now(),
	}, nil
}
