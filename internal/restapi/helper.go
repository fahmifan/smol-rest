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

type ServiceError int

const (
	ErrInternal                   ServiceError = 1000
	ErrPermissionDenied           ServiceError = 1001
	ErrInvalidArgument            ServiceError = 1002
	ErrNotFound                   ServiceError = 1003
	ErrMissingAuthorizationHeader ServiceError = 1004
	ErrUnauthorized               ServiceError = 1005
	ErrInvalidToken               ServiceError = 1006
)

func (s ServiceError) Error() string {
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
	Error string       `json:"error"`
	Code  ServiceError `json:"code"`
}

func jsonError(rw http.ResponseWriter, err error) {
	var svcErr ServiceError
	if svc, ok := err.(ServiceError); ok {
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
	// default ErrInternal
	default:
		log.Error().Err(err).Msg("")
		statusCode = http.StatusInternalServerError
	case ErrInvalidArgument:
		statusCode = http.StatusBadRequest
	case ErrPermissionDenied:
		statusCode = http.StatusForbidden
	case ErrNotFound:
		statusCode = http.StatusNotFound
	case ErrUnauthorized:
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
