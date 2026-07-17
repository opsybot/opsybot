package entity

import "errors"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

var ErrRoleInvalid = errors.New("role invalid")

func (r Role) Validate() error {
	switch r {
	case RoleAdmin, RoleMember:
		return nil
	default:
		return ErrRoleInvalid
	}
}
