package definitions

type SmolService interface {
	AddTodo(AddTodoRequest) Todo
	FindAllTodos(FindAllTodosFilter) Todos
	FindCurrentUser(Empty) User
	LogoutUser(Empty) Empty
}

type Empty struct{}

type User struct {
	ID    string
	Email string
	Role  string
}

type Todos struct {
	Todos []Todo
}

type FindAllTodosFilter struct {
	Page int
	Size int
}

type AddTodoRequest struct {
	Detail string
	Done   bool
}

type Todo struct {
	ID     string
	UserID string
	Done   bool
	Detail string
}
