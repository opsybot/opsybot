package entity

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

const (
	PasswordMinLength  = 12
	PasswordMaxLength  = 128
	NameMaxLength      = 80
	EmailMaxLength     = 120
	ArgonTime          = 2
	ArgonMemoryKiB     = 19456
	ArgonThreads       = 1
	ArgonKeyLength     = 32
	ArgonSaltLength    = 16
	LastActiveThrottle = time.Minute
)

type User struct {
	ID           string
	Email        string
	Name         string
	Timezone     string
	HasPassword  bool
	TOTPEnabled  bool
	LastActiveAt time.Time
	CreatedAt    time.Time
}

type NewUser struct {
	Email    string
	Name     string
	Password string
	Timezone string
}

type ProfileUpdate struct {
	Name     string
	Timezone string
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserEmailTaken      = errors.New("user email taken")
	ErrUserInvalidEmail    = errors.New("user invalid email")
	ErrUserInvalidName     = errors.New("user invalid name")
	ErrUserWeakPassword    = errors.New("user weak password")
	ErrUserInvalidTimezone = errors.New("user invalid timezone")
	ErrUserNoPassword      = errors.New("user has no password")
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > EmailMaxLength {
		return ErrUserInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Name != "" || addr.Address != email {
		return ErrUserInvalidEmail
	}
	return nil
}

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > NameMaxLength {
		return ErrUserInvalidName
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < PasswordMinLength || len(password) > PasswordMaxLength {
		return ErrUserWeakPassword
	}
	return nil
}

func ValidateTimezone(tz string) error {
	if strings.TrimSpace(tz) == "" {
		return ErrUserInvalidTimezone
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return ErrUserInvalidTimezone
	}
	return nil
}

func (n NewUser) Validate() error {
	if err := ValidateEmail(n.Email); err != nil {
		return err
	}
	if err := ValidateName(n.Name); err != nil {
		return err
	}
	if err := ValidatePassword(n.Password); err != nil {
		return err
	}
	return ValidateTimezone(n.Timezone)
}

func (p ProfileUpdate) Validate() error {
	if err := ValidateName(p.Name); err != nil {
		return err
	}
	return ValidateTimezone(p.Timezone)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
