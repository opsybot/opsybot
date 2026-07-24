package entity

import (
	"errors"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ServiceNameMaxLength        = 80
	ServiceDescriptionMaxLength = 280
	ServiceSlugMaxLength        = 40
	ServiceSlugMaxCandidates    = 100
)

func ServiceSlugCandidate(base string, n int) string {
	if base == "" {
		base = "service"
	}
	if n <= 1 {
		return base
	}
	suffix := "-" + strconv.Itoa(n)
	if max := ServiceSlugMaxLength - len(suffix); len(base) > max {
		base = base[:max]
	}
	return strings.TrimRight(base, "-") + suffix
}

type Service struct {
	ID          string
	WorkspaceID string
	Slug        string
	Name        string
	TeamID      string
	TeamSlug    string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewService struct {
	Name        string
	TeamSlug    string
	Description string
}

type ServiceUpdate struct {
	Name        string
	TeamSlug    string
	Description string
}

var (
	ErrServiceNotFound  = errors.New("service not found")
	ErrServiceSlugTaken = errors.New("service slug taken")
)

func (n NewService) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Name, validation.By(serviceNameField)),
		validation.Field(&n.Description, validation.By(serviceDescriptionField)),
	)
}

func (u ServiceUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(serviceNameField)),
		validation.Field(&u.Description, validation.By(serviceDescriptionField)),
	)
}
