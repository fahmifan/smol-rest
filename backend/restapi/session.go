package restapi

import (
	"context"
	"encoding/gob"

	"github.com/alexedwards/scs/v2"
	"github.com/fahmifan/smol/backend/model"
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

func NewSessionManager() *SessionManager {
	sess := scs.New()
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
