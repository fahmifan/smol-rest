package model

import "github.com/oklog/ulid/v2"

type Todo struct {
	ID     ulid.ULID
	UserID string
	Detail string
	Done   bool
}

func NewTodo(userID, detail string, done bool) Todo {
	return Todo{
		ID:     NewID(),
		UserID: userID,
		Detail: detail,
		Done:   done,
	}
}
