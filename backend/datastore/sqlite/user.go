package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fahmifan/smol/backend/model"
)

func (s *SQLite) SaveUser(ctx context.Context, user model.User) error {
	_, err := s.DB.ExecContext(ctx, `
	INSERT INTO 
	users (id, name, email, role) 
	VALUES (?, ?, ?, ?);
	`, user.ID.String(), user.Name, user.Email, user.Role)
	if err != nil {
		return fmt.Errorf("unable to insert users: %w", err)
	}
	return nil
}

var (
	ErrNotFound = errors.New("not found")
)

type sqlScanner interface {
	Scan(dest ...interface{}) error
}

// add space before & after column
const userRowColumn = `

	users.id,
	users.name,
	users.email,
	users.role

`

func userRowScan(s sqlScanner, u *model.User) error {
	return s.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Role,
	)
}

func (s *SQLite) FindUserByEmail(ctx context.Context, email string) (model.User, error) {
	query := `SELECT` + userRowColumn + `FROM users WHERE email = ?`
	row := s.DB.QueryRowContext(ctx, query, email)
	if err := row.Err(); err != nil {
		return model.User{}, fmt.Errorf("unable to find user by email: %w", err)
	}

	user := model.User{}
	err := userRowScan(row, &user)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("unable to scan user: %w", err)
	}

	return user, nil
}
