package model

import (
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

type User struct {
	ID        ulid.ULID
	Name      string
	Email     string
	Role      Role
	CreatedAt time.Time
	DeletedAt *time.Time
}

func NewUser(role Role, name, email string) (User, error) {
	switch role {
	case RoleGuest:
		return User{}, errors.New("cannot create user for Role Guest")
	}
	return User{
		ID:        NewID(),
		Name:      name,
		Role:      role,
		Email:     email,
		CreatedAt: time.Now(),
	}, nil
}
