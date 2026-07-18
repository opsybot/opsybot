package entity

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) Validate() error {
	return roleField(r)
}
