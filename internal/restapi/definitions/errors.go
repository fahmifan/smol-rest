package definitions

type ServiceError int

const (
	ErrInternal          ServiceError = 1000
	ErrPermissionDenined ServiceError = 1001
	ErrInvalidArgument   ServiceError = 1002
)

func (s ServiceError) Error() string {
	switch s {
	default: // ErrInternal
		return "internal"
	case ErrPermissionDenined:
		return "permission_denied"
	}
}
