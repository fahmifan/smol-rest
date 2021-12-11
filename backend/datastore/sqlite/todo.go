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
		todo.ID.String(), todo.UserID.String(), todo.Detail, todo.Done)
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

func (s *SQLite) FindAllUserTodos(ctx context.Context, userID string) ([]model.Todo, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT`+todoRowColumn+`FROM todos WHERE user_id = ?`, userID)
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
