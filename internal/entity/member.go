package entity

import (
	"errors"
	"time"
)

type MemberStatus string

const (
	MemberStatusInvited     MemberStatus = "invited"
	MemberStatusActive      MemberStatus = "active"
	MemberStatusDeactivated MemberStatus = "deactivated"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodSSO      AuthMethod = "sso"
	AuthMethodInvited  AuthMethod = "invited"
)

type Member struct {
	UserID        string
	WorkspaceID   string
	Name          string
	Email         string
	Role          Role
	Status        MemberStatus
	TOTPEnabled   bool
	HasPassword   bool
	HasSSO        bool
	LastActiveAt  time.Time
	JoinedAt      time.Time
	DeactivatedAt time.Time
	References    []MemberReference
}

func (m Member) Auth() AuthMethod {
	switch {
	case m.Status == MemberStatusInvited:
		return AuthMethodInvited
	case m.HasSSO:
		return AuthMethodSSO
	default:
		return AuthMethodPassword
	}
}

var (
	ErrMemberNotFound               = errors.New("member not found")
	ErrMemberAlreadyExists          = errors.New("member already exists")
	ErrMemberDeactivated            = errors.New("member deactivated")
	ErrMemberNotDeactivated         = errors.New("member not deactivated")
	ErrMemberLastAdmin              = errors.New("member last admin")
	ErrMemberReplacementsIncomplete = errors.New("member replacements incomplete")
	ErrMemberReplacementInvalid     = errors.New("member replacement invalid")
)
