package restapi

import (
	"context"

	"github.com/fahmifan/smol/backend/model"
	"github.com/fahmifan/smol/backend/restapi/generated"
	"github.com/rs/zerolog/log"
)

var _ generated.SmolService = &SmolService{}

type SmolService struct {
	*Server
}

type ServiceError int

const (
	ErrInternal          ServiceError = 1000
	ErrPermissionDenined ServiceError = 1001
)

func (s ServiceError) Error() string {
	switch s {
	default: // ErrInternal
		return "internal"
	case ErrPermissionDenined:
		return "permission_denied"
	}
}

func (g SmolService) AddTodo(ctx context.Context, r generated.AddTodoRequest) (*generated.Todo, error) {
	user := g.session.GetUser(ctx)
	if !user.Role.GrantedAny(model.Create_Todo) {
		return nil, ErrPermissionDenined
	}
	todo := model.NewTodo(
		user.UserID,
		r.Item,
		r.Done,
	)
	err := g.DataStore.SaveTodo(ctx, todo)
	if err != nil {
		log.Error().Err(err).Msg("AddTodo")
		return nil, ErrInternal
	}
	resp := &generated.Todo{
		ID:     todo.ID,
		UserID: todo.UserID,
		Done:   todo.Done,
		Detail: todo.Detail,
	}
	return resp, nil
}
