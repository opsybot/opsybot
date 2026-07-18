package entity

import (
	"errors"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	ErrUserNotFound   = errors.New("user not found")
	ErrUserEmailTaken = errors.New("user email taken")
	ErrUserNoPassword = errors.New("user has no password")
)

func ValidateEmail(email string) error {
	return validation.Validate(email, validation.By(emailField))
}

func ValidatePassword(password string) error {
	return validation.Validate(password, validation.By(passwordField))
}

func (n NewUser) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Email, validation.By(emailField)),
		validation.Field(&n.Name, validation.By(nameField)),
		validation.Field(&n.Password, validation.By(passwordField)),
		validation.Field(&n.Timezone, validation.By(timezoneField)),
	)
}

func (p ProfileUpdate) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.By(nameField)),
		validation.Field(&p.Timezone, validation.By(timezoneField)),
	)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
