package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/oklog/ulid/v2"
)

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// add space before & after column
const todoRowColumn = `

	todos.id,
	todos.user_id,
	todos.detail,
	todos.done

`

// add space before & after column
var todoColumns = []string{"id", "user_id", "detail", "done"}

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

func (s *Postgres) FindAllUserTodos(ctx context.Context, userID ulid.ULID, filter datastore.FindAllTodoFilter) (_ []model.Todo, count uint64, err error) {
	baseBuilder := sq.Select().From("todos").Where("user_id = ?", userID.String())
	queryCount, countArgs, err := baseBuilder.Column("COUNT(1)").ToSql()
	if err != nil {
		return
	}

	sqData := baseBuilder.Columns(todoColumns...).Limit(uint64(filter.GetSize())).OrderBy("id ASC")
	if filter.Cursor != "" {
		if filter.Backward {
			sqData = sqData.Where("id < ?", filter.Cursor).OrderBy("id DESC")
		} else {
			sqData = sqData.Where("id > ?", filter.Cursor).OrderBy("id ASC")
		}
	}

	queryData, args, err := sqData.ToSql()
	if err != nil {
		return
	}

	rowCount := s.DB.QueryRow(ctx, queryCount, countArgs...)
	err = rowCount.Scan(&count)
	if err != nil {
		err = fmt.Errorf("unable to scan todos count: %w", err)
		return
	}

	rows, err := s.DB.Query(ctx, queryData, args...)
	if err != nil {
		err = fmt.Errorf("unable to FindAllUserTodos: %w", err)
		return
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		todo := model.Todo{}
		if err = todoRowScan(rows, &todo); err != nil {
			err = fmt.Errorf("unable to scan todo: %w", err)
			return
		}
		todos = append(todos, todo)
	}

	return todos, count, nil
}
