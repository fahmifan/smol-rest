package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/jackc/pgx/v4"
	"github.com/oklog/ulid/v2"
)

// add space before & after column
const userRowColumn = `

	users.id,
	users.name,
	users.email,
	users.role

`

func userRowScan(s SqlScanner, u *model.User) error {
	return s.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Role,
	)
}

func (p *Postgres) SaveUser(ctx context.Context, user model.User) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO 
		"users" (id, name, email, role) 
		VALUES ($1, $2, $3, $4);`,
		user.ID.String(), user.Name, user.Email, user.Role,
	)
	return err
}

func (p *Postgres) FindUserByEmail(ctx context.Context, email string) (model.User, error) {
	query := `SELECT` + userRowColumn + `FROM users WHERE email = $1`
	row := p.DB.QueryRow(ctx, query, email)
	user := model.User{}
	err := userRowScan(row, &user)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, datastore.ErrNotFound
	}
	if err != nil {
		return user, fmt.Errorf("unable to scan: %w", err)
	}
	return user, nil
}

func (p *Postgres) FindUserByID(ctx context.Context, id ulid.ULID) (model.User, error) {
	query := `SELECT` + userRowColumn + `FROM users WHERE id = $1`
	row := p.DB.QueryRow(ctx, query, id.String())
	user := model.User{}
	err := userRowScan(row, &user)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, datastore.ErrNotFound
	}
	if err != nil {
		return user, fmt.Errorf("unable to scan: %w", err)
	}
	return user, nil
}
