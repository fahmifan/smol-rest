package restapi

import (
	"context"
	"database/sql"
	"encoding/gob"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/fahmifan/smol/backend/model"
	_ "github.com/mattn/go-sqlite3"
)

const userSessionKey = "user"

type UserSession struct {
	ID   string
	Role model.Role
}

func init() {
	gob.Register(UserSession{})
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

func (s *SessionManager) GetUser(ctx context.Context) *UserSession {
	userSess, _ := s.session.Get(ctx, userSessionKey).(*UserSession)
	return userSess
}

func (s *SessionManager) PutUser(ctx context.Context, user *UserSession) {
	s.session.Put(ctx, userSessionKey, user)
}
