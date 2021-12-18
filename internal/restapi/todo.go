package restapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/rs/zerolog/log"
)

func DecodeCursor(enc string) (dec string) {
	if enc == "" {
		return
	}
	bt, _ := base64.StdEncoding.DecodeString(enc)
	return string(bt)
}

func EncodeCursor(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

type PaginationResponse struct {
	Cursor   string `json:"cursor"`
	Backward bool   `json:"backward"`
	HasNext  bool   `json:"hasNext"`
	Count    uint64 `json:"count"`
	Size     uint64 `json:"size"`
}

func NewPaginationResponse(cursor string, backward bool, count uint64, size uint64, lenData int) PaginationResponse {
	return PaginationResponse{
		Cursor:   cursor,
		Backward: backward,
		Count:    count,
		HasNext:  lenData > 0,
	}
}

type PaginationRequest struct {
	Cursor   string `json:"cursor"`
	Backward bool   `json:"backward"`
	Size     uint64 `json:"size"`
}

type ResponseWithPaging struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination,omitempty"`
}

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
			jsonError(rw, err)
			return
		}

		ctx := r.Context()
		user := getUserFromCtx(ctx)
		if user.IsEmpty() {
			log.Debug().Msg(models.JSONS(user))
			jsonError(rw, ErrNotFound)
			return
		}

		todo := model.NewTodo(
			user.ID,
			req.Detail,
			req.Done,
		)
		err = s.DataStore.SaveTodo(ctx, todo)
		if err != nil {
			jsonError(rw, err)
			return
		}

		resp := Todo{
			ID:     todo.ID.String(),
			UserID: todo.UserID.String(),
			Done:   todo.Done,
			Detail: todo.Detail,
		}

		jsonOK(rw, resp)
	}
}

type FindAllTodosRequest struct {
	Pagination PaginationRequest `json:"pagination"`
}

// FindAllTodos ..
// @Summary find all todos
// @Description find all todos
// @ID FindAllTodos
// @Accept json
// @Produce json
// @Param user body FindAllTodosRequest true "find all todos request"
// @Success 200 {object} ResponseWithPaging
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
		todos, count, err := s.DataStore.FindAllUserTodos(ctx, user.ID, datastore.FindAllTodoFilter{
			Cursor:   DecodeCursor(req.Pagination.Cursor),
			Size:     req.Pagination.Size,
			Backward: req.Pagination.Backward,
		})
		if err != nil {
			jsonError(rw, err)
			return
		}

		var resTodos []Todo
		for _, todo := range todos {
			resTodos = append(resTodos, Todo{
				ID:     todo.ID.String(),
				UserID: todo.UserID.String(),
				Done:   todo.Done,
				Detail: todo.Detail,
			})
		}

		var cursor string
		if ntodo := len(todos); ntodo > 0 {
			if req.Pagination.Backward {
				cursor = EncodeCursor(todos[0].ID.String())
			} else {
				cursor = EncodeCursor(todos[ntodo-1].ID.String())
			}
		}

		res := ResponseWithPaging{
			Data: resTodos,
			Pagination: NewPaginationResponse(
				cursor,
				req.Pagination.Backward,
				count,
				uint64(req.Pagination.Size),
				len(resTodos),
			),
		}
		jsonOK(rw, res)
	}
}
