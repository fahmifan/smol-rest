package restapi

import (
	"context"

	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/restapi/gen"
	"github.com/rs/zerolog/log"
)

var _ gen.SmolService = &SmolService{}

func (g SmolService) AddTodo(ctx context.Context, r gen.AddTodoRequest) (*gen.Todo, error) {
	user := g.session.GetUser(ctx)
	if !user.Role.GrantedAny(model.Create_Todo) {
		return nil, ErrPermissionDenied
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
	resp := &gen.Todo{
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
		return nil, ErrPermissionDenied
	}

	todos, err := g.DataStore.FindAllUserTodos(ctx, sess.UserID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}

	var res []gen.Todo
	for _, todo := range todos {
		res = append(res, gen.Todo{
			ID:     todo.ID.String(),
			UserID: todo.UserID.String(),
			Done:   todo.Done,
			Detail: todo.Detail,
		})
	}
	return &gen.Todos{Todos: res}, nil
}
