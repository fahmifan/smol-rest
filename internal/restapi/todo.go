package restapi

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/rs/zerolog/log"
)

type AddTodoRequest struct {
	Detail string `json:"detail"`
	Done   bool   `json:"done"`
}

type Todo struct {
	ID     string `json:"id"`
	UserID string `json:"userID"`
	Done   bool   `json:"done"`
	Detail string `json:"detail"`
}

type FindAllTodoRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func NewFindAllTodoRequest(query url.Values) FindAllTodoRequest {
	page := models.StringToInt(query.Get("page"))
	size := models.StringToInt(query.Get("size"))

	req := FindAllTodoRequest{}
	if page < 1 {
		req.Page = 1
	}
	if size < 1 || size > 25 {
		req.Size = 25
	}
	return req
}

func (f *FindAllTodoRequest) ParseQuery(query url.Values) {
	f.Page = models.StringToInt(query.Get("page"))
	f.Size = models.StringToInt(query.Get("size"))
}

// Create Todo godoc
// @Summary create a new todo
// @Description currently it only support one session per user
// @ID Login
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
			httpError(rw, err)
			return
		}

		ctx := r.Context()
		user := getUserFromCtx(ctx)
		if user.IsEmpty() {
			log.Debug().Msg(models.JSONS(user))
			httpError(rw, ErrNotFound)
			return
		}

		todo := model.NewTodo(
			user.ID,
			req.Detail,
			req.Done,
		)
		err = s.DataStore.SaveTodo(ctx, todo)
		if err != nil {
			httpError(rw, err)
			return
		}

		resp := Todo{
			ID:     todo.ID.String(),
			UserID: todo.UserID.String(),
			Done:   todo.Done,
			Detail: todo.Detail,
		}

		httpOK(rw, resp)
	}
}

func (s *Server) HandleFindAllTodos() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user := getUserFromCtx(ctx)
		todos, err := s.DataStore.FindAllUserTodos(ctx, user.ID)
		if err != nil {
			httpError(rw, err)
			return
		}

		var res []Todo
		for _, todo := range todos {
			res = append(res, Todo{
				ID:     todo.ID.String(),
				UserID: todo.UserID.String(),
				Done:   todo.Done,
				Detail: todo.Detail,
			})
		}

		httpOK(rw, res)
	}
}
