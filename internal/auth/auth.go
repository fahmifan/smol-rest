package auth

import (
	"errors"

	"github.com/fahmifan/flycasbin/acl"
)

var _ACL *acl.ACL

func SetACL(a *acl.ACL) {
	_ACL = a
}

type Role = acl.Role

const (
	RoleAdmin = Role("admin")
	RoleUser  = Role("user")
	RoleGuest = Role("guest")
)

type Resource = acl.Resource

const (
	Dashboard Resource = "dashboard"
	Todo      Resource = "todo"
)

type Action = acl.Action

const (
	View     Action = "view"
	ViewAny  Action = "view_any"
	ViewSelf Action = "view_self"
	Create   Action = "create"
)

func _P(r Role, act Action, rsc Resource) acl.Policy {
	return acl.Policy{
		Role:     r,
		Resource: rsc,
		Action:   act,
	}
}

var Policies = []acl.Policy{
	_P(RoleAdmin, View, Dashboard),
	_P(RoleAdmin, Create, Todo),
	_P(RoleAdmin, ViewAny, Todo),

	_P(RoleUser, View, Dashboard),
	_P(RoleUser, Create, Todo),
	_P(RoleUser, ViewSelf, Todo),
}

type Permission struct {
	Action   Action
	Resource Resource
}

func Perm(act Action, rsc Resource) Permission {
	return Permission{Action: act, Resource: rsc}
}

var ErrPermissionDenied = errors.New("permission denied")
var ErrACLNotSet = errors.New("acl not set")

// GrantedAny check if the Role is granted any permissions.
// If a permission is given, returns nil
func GrantedAny(role Role, perms ...Permission) error {
	if _ACL == nil {
		return ErrACLNotSet
	}

	for _, perm := range perms {
		err := _ACL.Can(acl.Role(role), acl.Action(perm.Action), acl.Resource(perm.Resource))
		if errors.Is(err, acl.ErrPermissionDenied) {
			return ErrPermissionDenied
		}
		if err != nil {
			return err
		}
	}

	return nil
}
