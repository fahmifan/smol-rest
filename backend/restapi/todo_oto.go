package restapi

import (
	"context"

	"github.com/fahmifan/smol/backend/model"
	"github.com/fahmifan/smol/backend/restapi/gen"
	generated "github.com/fahmifan/smol/backend/restapi/gen"
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
	ErrInvalidArgument   ServiceError = 1002
)

func (s ServiceError) Error() string {
	switch s {
	default: // ErrInternal
		return "internal"
	case ErrPermissionDenined:
		return "permission_denied"
	case ErrInvalidArgument:
		return "invalid_argument"
	}
}

func (g SmolService) AddTodo(ctx context.Context, r generated.AddTodoRequest) (*generated.Todo, error) {
	user := g.session.GetUser(ctx)
	if !user.Role.GrantedAny(model.Create_Todo) {
		return nil, ErrPermissionDenined
	}
	todo := model.NewTodo(
		user.UserID,
		r.Detail,
		r.Done,
	)
	err := g.DataStore.SaveTodo(ctx, todo)
	if err != nil {
		log.Error().Err(err).Msg("AddTodo")
		return nil, ErrInternal
	}
	resp := &generated.Todo{
		ID:     todo.ID.String(),
		UserID: todo.UserID.String(),
		Done:   todo.Done,
		Detail: todo.Detail,
	}
	return resp, nil
}

func (g SmolService) FindAllTodos(ctx context.Context, r gen.FindAllTodosFilter) (*gen.Todos, error) {
	sess := g.session.GetUser(ctx)
	if !sess.Role.GrantedAny(model.View_AllSelfTodo) {
		return nil, ErrPermissionDenined
	}

	todos, err := g.DataStore.FindAllUserTodos(ctx, sess.UserID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}

	var res []gen.Todo
	for _, todo := range todos {
		res = append(res, generated.Todo{
			ID:     todo.ID.String(),
			UserID: todo.UserID.String(),
			Done:   todo.Done,
			Detail: todo.Detail,
		})
	}
	return &generated.Todos{Todos: res}, nil
}
