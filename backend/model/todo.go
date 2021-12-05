package model

type Todo struct {
	ID     string
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
