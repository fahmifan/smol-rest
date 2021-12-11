package model

import "github.com/oklog/ulid/v2"

type Todo struct {
	ID     ulid.ULID
	UserID ulid.ULID
	Detail string
	Done   bool
}

func NewTodo(userID, detail string, done bool) Todo {
	uid, _ := ulid.Parse(userID)
	return Todo{
		ID:     NewID(),
		UserID: uid,
		Detail: detail,
		Done:   done,
	}
}
