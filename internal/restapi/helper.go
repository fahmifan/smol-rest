package restapi

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/fahmifan/smol/internal/usecase"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
)

type SmolService struct {
	*Server
}

type ErrorCode int

const (
	ErrInternal                   ErrorCode = 1000
	ErrPermissionDenied           ErrorCode = 1001
	ErrInvalidArgument            ErrorCode = 1002
	ErrNotFound                   ErrorCode = 1003
	ErrMissingAuthorizationHeader ErrorCode = 1004
	ErrUnauthorized               ErrorCode = 1005
	ErrInvalidToken               ErrorCode = 1006
	ErrRefreshTokenExpired        ErrorCode = 1007
)

func (s ErrorCode) Error() string {
	switch s {
	default: // ErrInternal
		return "internal"
	case ErrPermissionDenied:
		return "permission_denied"
	case ErrInvalidArgument:
		return "invalid_argument"
	case ErrNotFound:
		return "not_found"
	case ErrMissingAuthorizationHeader:
		return "missing_authorization"
	case ErrUnauthorized:
		return "unauthorized"
	case ErrInvalidToken:
		return "invalid_token"
	case ErrRefreshTokenExpired:
		return "Refresh_token_expired"
	}
}

type Map map[string]interface{}

func writeJSON(rw http.ResponseWriter, status int, body interface{}) {
	rw.WriteHeader(status)
	rw.Write(models.JSON(body))
}

func jsonOK(rw http.ResponseWriter, res interface{}) {
	writeJSON(rw, http.StatusOK, res)
}

type ErrorResponse struct {
	Error   string    `json:"error"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message,omitempty"`
}

// jsonError write error with a message
// the variadic msg param will only used the index 0 value
func jsonError(rw http.ResponseWriter, err error, msgs ...string) {
	var svcErr ErrorCode
	if svc, ok := err.(ErrorCode); ok {
		svcErr = svc
	} else {
		switch err {
		default:
			svcErr = ErrInternal
		case usecase.ErrInvalidArgument:
			svcErr = ErrInvalidArgument
		case datastore.ErrNotFound, pgx.ErrNoRows, sql.ErrNoRows:
			svcErr = ErrNotFound
		}
	}

	var statusCode int
	switch svcErr {
	default:
		log.Error().Int("code", int(svcErr)).Err(err).Msg("unknown code")
		statusCode = http.StatusInternalServerError
	case ErrInternal:
		log.Error().Err(err).Msg("")
		statusCode = http.StatusInternalServerError
	case ErrInvalidArgument, ErrInvalidToken, ErrRefreshTokenExpired:
		statusCode = http.StatusBadRequest
	case ErrPermissionDenied:
		statusCode = http.StatusForbidden
	case ErrNotFound:
		statusCode = http.StatusNotFound
	case ErrUnauthorized, ErrMissingAuthorizationHeader:
		statusCode = http.StatusUnauthorized
	}

	var msg string
	if len(msgs) > 0 {
		msg = msgs[0]
	}

	writeJSON(rw, statusCode, ErrorResponse{
		Error:   svcErr.Error(),
		Code:    svcErr,
		Message: msg,
	})
}

type ctxKey string

const userSessionCtxKey ctxKey = "user_session"

func getUserFromCtx(c context.Context) usecase.UserToken {
	res := c.Value(userSessionCtxKey)
	if res == nil {
		return usecase.UserToken{}
	}
	if val, ok := res.(usecase.UserToken); ok {
		return val
	}

	return usecase.UserToken{}
}

func setUserToCtx(c context.Context, user usecase.UserToken) context.Context {
	return context.WithValue(c, userSessionCtxKey, user)
}
