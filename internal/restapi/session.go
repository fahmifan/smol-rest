package restapi

import (
	"context"
	"database/sql"
	"encoding/gob"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/fahmifan/smol/internal/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
)

const userSessionKey = "user"

type Session struct {
	UserID    ulid.ULID
	Role      model.Role
	ExpiredAt time.Time
}

func init() {
	gob.Register(Session{})
}

type SessionManager struct {
	session *scs.SessionManager
}

func NewSessionManager(db *sql.DB) *SessionManager {
	sess := scs.New()
	sess.Store = sqlite3store.New(db)
	return &SessionManager{
		session: sess,
	}
}

func (s *SessionManager) GetUser(ctx context.Context) Session {
	val := s.session.Get(ctx, userSessionKey)
	userSess, ok := val.(Session)
	if !ok {
		return Session{Role: model.RoleGuest}
	}
	return userSess
}

func (s *SessionManager) PutUser(ctx context.Context, user *Session) {
	s.session.Put(ctx, userSessionKey, user)
}

func (s *SessionManager) PopUser(ctx context.Context) {
	_ = s.session.Pop(ctx, userSessionKey)
}
