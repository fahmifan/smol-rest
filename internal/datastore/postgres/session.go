package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/jackc/pgx/v4"
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

func sessionRowScan(s SqlScanner, sess *model.Session) error {
	return s.Scan(
		&sess.ID,
		&sess.UserID,
		&sess.RefreshToken,
		&sess.RefreshTokenExpiredAt,
		&sess.CreatedAt,
	)
}

func (p *Postgres) CreateSession(ctx context.Context, sess model.Session) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO "sessions" 
		(id, user_id, refresh_token_expired_at, refresh_token, created_at) VALUES
		($1,		$2,						 $3, 			$4,			$5)`,
		sess.ID.String(), sess.UserID.String(), sess.RefreshTokenExpiredAt, sess.RefreshToken, sess.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("unable to insert session: %w", err)
	}
	return nil
}

func (p *Postgres) FindSessionByRefreshToken(ctx context.Context, token string) (model.Session, error) {
	sess := model.Session{}
	row := p.DB.QueryRow(ctx, `SELECT`+sessionRowColumn+`FROM "sessions" WHERE refresh_token = $1`, token)
	err := sessionRowScan(row, &sess)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Session{}, datastore.ErrNotFound
	}
	return sess, nil
}

func (p *Postgres) DeleteSessionByUserID(ctx context.Context, userID ulid.ULID) error {
	_, err := p.DB.Exec(ctx, `DELETE FROM "sessions" WHERE user_id = $1`, userID)
	return fmt.Errorf("unable to delete session: %w", err)
}
