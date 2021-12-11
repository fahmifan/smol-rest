package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fahmifan/smol/backend/model"
)

// add space before & after column
const todoRowColumn = `

	todos.id,
	todos.user_id,
	todos.detail,
	todos.done

`

func todoRowScan(s sqlScanner, t *model.Todo) error {
	return s.Scan(
		&t.ID,
		&t.UserID,
		&t.Detail,
		&t.Done,
	)
}

func (s *SQLite) SaveTodo(ctx context.Context, todo model.Todo) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO todos 
		(id, user_id, detail, done) VALUES
		( ?, 		?, 		?, 	 ?);`,
		todo.ID.String(), todo.UserID, todo.Detail, todo.Done)
	if err != nil {
		return fmt.Errorf("unable to saveTodo: %w", err)
	}
	return nil
}

func (s *SQLite) FindTodoByID(ctx context.Context, id string) (model.Todo, error) {
	row := s.DB.QueryRowContext(ctx, `SELECT`+todoRowColumn+`FROM todos WHERE id = ?`, id)
	if err := row.Err(); err != nil {
		return model.Todo{}, fmt.Errorf("unable to FindTodoByID: %w", err)
	}

	todo := model.Todo{}
	err := todoRowScan(row, &todo)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, ErrNotFound
	}
	if err != nil {
		return todo, fmt.Errorf("unable to scan todo: %w", err)
	}

	return todo, nil
}
