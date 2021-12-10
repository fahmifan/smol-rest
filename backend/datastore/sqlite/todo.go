package sqlite

import (
	"context"
	"fmt"

	"github.com/fahmifan/smol/backend/model"
)

func (s *SQLite) SaveTodo(ctx context.Context, todo model.Todo) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO todos 
		(id, user_id, detail, done) VALUES
		( ?, 		?, 		?, 	 ?);`,
		todo.ID, todo.UserID, todo.Detail, todo.Done)
	if err != nil {
		return fmt.Errorf("unable to saveTodo: %w", err)
	}
	return nil
}
