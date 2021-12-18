package restapi

import (
	"context"
	"net/http"

	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/model/models"
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
	Error string    `json:"error"`
	Code  ErrorCode `json:"code"`
}

func jsonError(rw http.ResponseWriter, err error) {
	var svcErr ErrorCode
	if svc, ok := err.(ErrorCode); ok {
		svcErr = svc
	} else {
		switch err {
		default:
			svcErr = ErrInternal
		case model.ErrInvalidArgument:
			svcErr = ErrInvalidArgument
		case datastore.ErrNotFound:
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

	writeJSON(rw, statusCode, ErrorResponse{Error: svcErr.Error(), Code: svcErr})
}

type ctxKey string

const userSessionCtxKey ctxKey = "user_session"

func getUserFromCtx(c context.Context) model.User {
	res := c.Value(userSessionCtxKey)
	if res == nil {
		return model.User{}
	}
	if val, ok := res.(model.User); ok {
		return val
	}

	return model.User{}
}

func setUserToCtx(c context.Context, user model.User) context.Context {
	return context.WithValue(c, userSessionCtxKey, user)
}
