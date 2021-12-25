package usecase

import (
	"context"
	"strings"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
)

type Todoer struct {
	Queries *sqlcpg.Queries
}

func (t *Todoer) FindAllTodos(ctx context.Context, userID string, filter datastore.FindAllTodoFilter) (todos []sqlcpg.Todo, count int64, err error) {
	count, err = t.Queries.CountAllTodos(ctx, userID)
	if err != nil {
		return
	}

	if strings.TrimSpace(filter.Cursor) == "" {
		todos, err = t.Queries.FindAllUserTodos(ctx, sqlcpg.FindAllUserTodosParams{
			UserID: userID,
			Size:   int32(filter.Size),
		})
		if err != nil {
			return
		}
		return
	}

	if filter.Backward {
		todos, err = t.Queries.FindAllUserTodosAsc(ctx, sqlcpg.FindAllUserTodosAscParams{
			UserID: userID,
			Cursor: filter.Cursor,
			Size:   int32(filter.Size),
		})
		if err != nil {
			return
		}
		return
	}

	todos, err = t.Queries.FindAllUserTodosDesc(ctx, sqlcpg.FindAllUserTodosDescParams{
		UserID: userID,
		Cursor: filter.Cursor,
		Size:   int32(filter.Size),
	})
	if err != nil {
		return
	}
	return
}
