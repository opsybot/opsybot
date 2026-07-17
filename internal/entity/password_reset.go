package entity

import (
	"errors"
	"time"
)

const (
	PasswordResetTTL         = time.Hour
	PasswordResetTokenLength = 32
)

type PasswordReset struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	UsedAt    time.Time
	CreatedAt time.Time
}

func (p PasswordReset) Usable() bool {
	return p.UsedAt.IsZero() && time.Now().Before(p.ExpiresAt)
}

var (
	ErrPasswordResetNotFound = errors.New("password reset not found")
	ErrPasswordResetInvalid  = errors.New("password reset invalid")
)
