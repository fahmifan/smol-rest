package definitions

type SmolService interface {
	AddTodo(AddTodoRequest) Todo
}

type AddTodoRequest struct {
	Item string
	Done bool
}

type Todo struct {
	ID     string
	UserID string
	Done   bool
	Detail string
}
