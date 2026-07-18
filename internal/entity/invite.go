package entity

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRevoked  InviteStatus = "revoked"
)

const (
	InviteTTL         = 14 * 24 * time.Hour
	InviteTokenLength = 32
)

type Invite struct {
	ID            string
	WorkspaceID   string
	UserID        string
	Email         string
	InvitedBy     string
	InvitedByName string
	WorkspaceName string
	WorkspaceSlug string
	Role          Role
	Status        InviteStatus
	ExpiresAt     time.Time
	AcceptedAt    time.Time
	CreatedAt     time.Time
}

type AcceptInvite struct {
	Token    string
	Name     string
	Password string
	Timezone string
}

type AcceptResult struct {
	Invite  Invite
	Session Session
	Token   string
	User    User
}

func (i Invite) Expired() bool {
	return i.Status == InviteStatusPending && time.Now().After(i.ExpiresAt)
}

var (
	ErrInviteNotFound        = errors.New("invite not found")
	ErrInviteExpired         = errors.New("invite expired")
	ErrInviteAlreadyAccepted = errors.New("invite already accepted")
	ErrInviteRevoked         = errors.New("invite revoked")
	ErrInvitePending         = errors.New("invite pending")
)

func (a AcceptInvite) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Name, validation.By(nameField)),
		validation.Field(&a.Password, validation.By(passwordField)),
		validation.Field(&a.Timezone, validation.By(timezoneField)),
	)
}
