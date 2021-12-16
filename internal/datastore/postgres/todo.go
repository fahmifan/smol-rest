package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/oklog/ulid/v2"
)

// add space before & after column
const todoRowColumn = `

	todos.id,
	todos.user_id,
	todos.detail,
	todos.done

`

func todoRowScan(s SqlScanner, t *model.Todo) error {
	return s.Scan(
		&t.ID,
		&t.UserID,
		&t.Detail,
		&t.Done,
	)
}

func (s *Postgres) SaveTodo(ctx context.Context, todo model.Todo) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO todos 
		(id, 	user_id, detail,  done) VALUES
		( $1,		 $2, 	 $3, 	$4);`,
		todo.ID.String(), todo.UserID.String(), todo.Detail, todo.Done)
	if err != nil {
		return fmt.Errorf("unable to saveTodo: %w", err)
	}
	return nil
}

func (s *Postgres) FindTodoByID(ctx context.Context, id string) (model.Todo, error) {
	row := s.DB.QueryRow(ctx, `SELECT`+todoRowColumn+`FROM todos WHERE id = ?`, id)
	todo := model.Todo{}
	err := todoRowScan(row, &todo)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, datastore.ErrNotFound
	}
	if err != nil {
		return todo, fmt.Errorf("unable to scan todo: %w", err)
	}

	return todo, nil
}

func (s *Postgres) FindAllUserTodos(ctx context.Context, userID ulid.ULID) ([]model.Todo, error) {
	rows, err := s.DB.Query(ctx, `SELECT`+todoRowColumn+`FROM todos WHERE user_id = $1`, userID.String())
	if err != nil {
		return nil, fmt.Errorf("unable to FindAllUserTodos: %w", err)
	}

	var todos []model.Todo
	for rows.Next() {
		todo := model.Todo{}
		if err = todoRowScan(rows, &todo); err != nil {
			return nil, fmt.Errorf("unable to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	return todos, nil
}
