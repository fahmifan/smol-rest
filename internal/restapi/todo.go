package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/rs/zerolog/log"
)

type Todo struct {
	ID     string `json:"id"`
	UserID string `json:"userID"`
	Done   bool   `json:"done"`
	Detail string `json:"detail"`
}

type AddTodoRequest struct {
	Detail string `json:"detail"`
	Done   bool   `json:"done"`
}

type FindAllTodosRequest struct {
	Pagination PaginationRequest `json:"pagination"`
}

// CreateTodo ..
// @Summary create a new todo
// @Description currently it only support one session per user
// @ID CreateTodo
// @Accept json
// @Produce json
// @Param user body AddTodoRequest true "add todo request"
// @Success 200 {object} Todo
// @Failure 400 {object} ErrorResponse
// @Router /api/todos [post]
func (s *Server) HandleCreateTodo() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := AddTodoRequest{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(rw, err)
			return
		}

		ctx := r.Context()
		user := getUserFromCtx(ctx)
		if user.ID == "" {
			log.Debug().Msg(models.JSONS(user))
			jsonError(rw, ErrNotFound)
			return
		}

		todo, err := s.Queries.SaveTodo(ctx, sqlcpg.SaveTodoParams{
			ID:     model.NewID().String(),
			Detail: req.Detail,
			UserID: user.ID,
			Done:   req.Done,
		})
		if err != nil {
			jsonError(rw, err)
			return
		}

		resp := Todo{
			ID:     todo.ID,
			UserID: todo.UserID,
			Done:   todo.Done,
			Detail: todo.Detail,
		}

		jsonOK(rw, resp)
	}
}

// FindAllTodos ..
// @Summary find all todos
// @Description find all todos
// @ID FindAllTodos
// @Accept json
// @Produce json
// @Param pagination body FindAllTodosRequest true "find all todos request"
// @Success 200 {object} ResponseWithPagination
// @Failure 400 {object} ErrorResponse
// @Router /api/todos [get]
func (s *Server) HandleFindAllTodos() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req := FindAllTodosRequest{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(rw, err)
			return
		}

		if req.Pagination.Backward && req.Pagination.Cursor == "" {
			jsonError(rw, ErrInvalidArgument, "cannot backward without cursor")
			return
		}

		user := getUserFromCtx(ctx)
		todos, count, err := s.Todoer.FindAllTodos(ctx, user.ID, datastore.FindAllTodoFilter{
			Cursor:   req.Pagination.Cursor,
			Backward: req.Pagination.Backward,
			Size:     uint64(req.Pagination.Size),
		})
		if err != nil {
			jsonError(rw, err)
			return
		}

		var resTodos []Todo
		for _, todo := range todos {
			resTodos = append(resTodos, Todo{
				ID:     todo.ID,
				UserID: todo.UserID,
				Done:   todo.Done,
				Detail: todo.Detail,
			})
		}

		var cursor string
		if ntodo := len(todos); ntodo > 0 {
			if req.Pagination.Backward {
				cursor = EncodeCursor(todos[0].ID)
			} else {
				cursor = EncodeCursor(todos[ntodo-1].ID)
			}
		}

		res := ResponseWithPagination{
			Data: resTodos,
			Pagination: NewPaginationResponse(
				cursor,
				req.Pagination.Backward,
				count,
				req.Pagination.Size,
				len(resTodos),
			),
		}
		jsonOK(rw, res)
	}
}
