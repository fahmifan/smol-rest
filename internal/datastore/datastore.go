package datastore

import (
	"context"
	"errors"

	"github.com/fahmifan/smol/internal/model"
	"github.com/oklog/ulid/v2"
)

var (
	ErrNotFound = errors.New("not found")
)

type UserReadWriter interface {
	SaveUser(ctx context.Context, user model.User) error
	FindUserByEmail(ctx context.Context, email string) (model.User, error)
	FindUserByID(ctx context.Context, id ulid.ULID) (model.User, error)
}

type SessionReadWriter interface {
	CreateSession(ctx context.Context, sess model.Session) error
	FindSessionByRefreshToken(ctx context.Context, token string) (model.Session, error)
	DeleteSessionByUserID(ctx context.Context, userID ulid.ULID) error
}

type FindAllTodoFilter struct {
	Cursor   string
	Backward bool
	Size     uint64
}

const MaxSize = 25

func (f *FindAllTodoFilter) GetSize() uint64 {
	if f.Size == 0 || f.Size > MaxSize {
		f.Size = MaxSize
	}

	return f.Size
}

type TodoReadWriter interface {
	SaveTodo(ctx context.Context, todo model.Todo) error
	FindTodoByID(ctx context.Context, id string) (model.Todo, error)
	FindAllUserTodos(ctx context.Context, userID ulid.ULID, filter FindAllTodoFilter) (todos []model.Todo, count uint64, err error)
}

type DataStore interface {
	UserReadWriter
	SessionReadWriter
	TodoReadWriter
}
