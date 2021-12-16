package restapi

import (
	"context"
	"net/http"

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

func httpOK(rw http.ResponseWriter, res interface{}) {
	writeJSON(rw, http.StatusOK, res)
}

func httpError(rw http.ResponseWriter, err error, svcErr ServiceError) {
	type ErrRes struct {
		Error string       `json:"error"`
		Code  ServiceError `json:"code"`
	}
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	var statusCode int
	switch svcErr {
	// default ErrInternal
	default:
		statusCode = http.StatusInternalServerError
	case ErrInvalidArgument:
		statusCode = http.StatusBadRequest
	case ErrPermissionDenied:
		statusCode = http.StatusUnauthorized
	case ErrNotFound:
		statusCode = http.StatusNotFound
	}

	writeJSON(rw, statusCode, ErrRes{Error: svcErr.Error(), Code: svcErr})
}

func getUserFromCtx(c context.Context) model.User {
	res := c.Value(userSessionKey)
	if val, ok := res.(model.User); ok {
		return val
	}

	return model.User{}
}

func setUserToCtx(c context.Context, user model.User) {
	context.WithValue(c, userSessionKey, user)
}
