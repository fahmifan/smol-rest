package model

import (
	"log"
	"sync"

	"github.com/fahmifan/smol/backend/model/models"
)

// Role ..
type Role int

// String ..
func (u Role) String() string {
	switch u {
	case RoleAdmin:
		return "admin"
	case RoleUser:
		return "user"
	case RoleGuest:
		return "guest"
	default:
		return ""
	}
}

// ParseRole ..
func ParseRole(s string) Role {
	switch s {
	case "admin":
		return RoleAdmin
	case "user":
		return RoleUser
	default:
		return RoleGuest
	}
}

// roles ..
const (
	RoleAdmin = Role(1)
	RoleUser  = Role(2)
	RoleGuest = Role(3)
)

type Permission int

const (
	View_Dashboard Permission = iota
	Create_Todo
	View_AllSelfTodo
)

var policy = map[Role][]Permission{
	RoleAdmin: {
		View_Dashboard,
		Create_Todo,
		View_AllSelfTodo,
	},
	RoleUser: {
		View_Dashboard,
		Create_Todo,
		View_AllSelfTodo,
	},
	RoleGuest: {},
}

// GrantedAny check if the Role is granted any permissions.
// If no permission given, it returns true
func (r Role) GrantedAny(perm ...Permission) bool {
	if len(perm) == 0 {
		return true
	}
	for _, p := range perm {
		if r.granted(p) {
			return true
		}
	}

	return false
}

// Granted check if role is granted with a permission
func (r Role) granted(perm Permission) bool {
	role, ok := cachePolicy[r]
	if !ok {
		return false
	}

	log.Print(models.JSONS(cachePolicy))
	return role[perm]
}

var cachePolicy map[Role]map[Permission]bool

var onceCachePolicy sync.Once

func init() {
	onceCachePolicy.Do(func() {
		cachePolicy = make(map[Role]map[Permission]bool)
		for role, perms := range policy {
			for _, perm := range perms {
				if cachePolicy[role] == nil {
					cachePolicy[role] = map[Permission]bool{}
				}
				cachePolicy[role][perm] = true
			}
		}
	})
}
