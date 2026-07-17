package entity

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	SlugPattern            = `^[a-z][a-z0-9-]{0,39}$`
	WorkspaceNameMaxLength = 80
)

var slugRe = regexp.MustCompile(SlugPattern)

var WorkspaceReservedSlugs = []string{
	"login", "signup", "setup", "invite", "forgot-password", "reset-password",
	"two-factor", "sso-error", "api", "assets",
}

type Workspace struct {
	ID          string
	Slug        string
	Name        string
	Timezone    string
	Environment string
	CreatedBy   string
	CreatedAt   time.Time
}

type NewWorkspace struct {
	Slug        string
	Name        string
	Timezone    string
	Environment string
}

type WorkspaceUpdate struct {
	Name        string
	Timezone    string
	Environment string
}

type Setup struct {
	UserName      string
	Email         string
	Password      string
	WorkspaceName string
	WorkspaceSlug string
	Timezone      string
}

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrWorkspaceSlugTaken    = errors.New("workspace slug taken")
	ErrWorkspaceSlugInvalid  = errors.New("workspace slug invalid")
	ErrWorkspaceSlugReserved = errors.New("workspace slug reserved")
	ErrWorkspaceNameInvalid  = errors.New("workspace name invalid")
	ErrSetupAlreadyDone      = errors.New("setup already done")
	ErrNotMember             = errors.New("not a workspace member")
)

func ValidateSlug(slug string, reserved []string) error {
	if !slugRe.MatchString(slug) {
		return ErrWorkspaceSlugInvalid
	}
	if slices.Contains(reserved, slug) {
		return ErrWorkspaceSlugReserved
	}
	return nil
}

func ValidateWorkspaceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > WorkspaceNameMaxLength {
		return ErrWorkspaceNameInvalid
	}
	return nil
}

func (n NewWorkspace) Validate() error {
	if err := ValidateSlug(n.Slug, WorkspaceReservedSlugs); err != nil {
		return err
	}
	if err := ValidateWorkspaceName(n.Name); err != nil {
		return err
	}
	return ValidateTimezone(n.Timezone)
}

func (u WorkspaceUpdate) Validate() error {
	if err := ValidateWorkspaceName(u.Name); err != nil {
		return err
	}
	return ValidateTimezone(u.Timezone)
}

func (s Setup) Validate() error {
	if err := ValidateName(s.UserName); err != nil {
		return err
	}
	if err := ValidateEmail(s.Email); err != nil {
		return err
	}
	if err := ValidatePassword(s.Password); err != nil {
		return err
	}
	if err := ValidateWorkspaceName(s.WorkspaceName); err != nil {
		return err
	}
	if err := ValidateSlug(s.WorkspaceSlug, WorkspaceReservedSlugs); err != nil {
		return err
	}
	return ValidateTimezone(s.Timezone)
}

func Slugify(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	if slug == "" || !(slug[0] >= 'a' && slug[0] <= 'z') {
		slug = "w" + slug
	}
	return slug
}
