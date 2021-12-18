package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fahmifan/smol/internal/model"
	"github.com/oklog/ulid/v2"
)

// add space before & after column
const sessionRowColumn = `

	sessions.id,
	sessions.user_id,
	sessions.refresh_token,
	sessions.refresh_token_expired_at,
	sessions.created_at

`

func sessionRowScan(s sqlScanner, sess *model.Session) error {
	return s.Scan(
		&sess.ID,
		&sess.UserID,
		&sess.RefreshToken,
		&sess.RefreshTokenExpiredAt,
		&sess.CreatedAt,
	)
}

func (s *SQLite) CreateSession(ctx context.Context, sess model.Session) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO "sessions" 
		(id, user_id, refresh_token_expired_at, refresh_token, created_at) VALUES
		(?,			?,						 ?, 			?,			?)`,
		sess.ID, sess.UserID, sess.RefreshTokenExpiredAt, sess.RefreshToken, sess.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("unable to insert session: %w", err)
	}

	return nil
}

func (s *SQLite) FindSessionByRefreshToken(ctx context.Context, token string) (model.Session, error) {
	q := `SELECT` + sessionRowColumn + `FROM sessions WHERE refresh_token = ? AND refresh_token_expired_at > ?`
	res := s.DB.QueryRowContext(ctx, q, token, time.Now())
	err := res.Err()
	if err != nil {
		return model.Session{}, fmt.Errorf("unable to find session by refresh token")
	}

	sess := model.Session{}
	err = sessionRowScan(res, &sess)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Session{}, ErrNotFound
	}
	if err != nil {
		return sess, fmt.Errorf("unable to scan session: %w", err)
	}

	return sess, nil
}

func (s *SQLite) DeleteSessionByUserID(ctx context.Context, userID ulid.ULID) error {
	q := `DELETE FROM "sessions" WHERE user_id = ?`
	_, err := s.DB.ExecContext(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("unable to find session by refresh token")
	}

	return nil
}
