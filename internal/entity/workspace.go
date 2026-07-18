package entity

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	SlugPattern                = `^[a-z][a-z0-9-]{0,39}$`
	SlugMaxLength              = 40
	WorkspaceNameMaxLength     = 80
	WorkspaceSlugMaxCandidates = 100
	SlugSuffixMin              = 1000
	SlugSuffixSpan             = 9000
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

type Signup struct {
	UserName      string
	Email         string
	Password      string
	WorkspaceName string
	WorkspaceSlug string
	Timezone      string
}

func (s Signup) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.UserName, validation.By(nameField)),
		validation.Field(&s.Email, validation.By(emailField)),
		validation.Field(&s.Password, validation.By(passwordField)),
		validation.Field(&s.WorkspaceName, validation.By(workspaceNameField)),
		validation.Field(&s.WorkspaceSlug, validation.By(slugField)),
		validation.Field(&s.Timezone, validation.By(timezoneField)),
	)
}

func WorkspaceSlugCandidate(base string, n int) string {
	if n <= 1 {
		return base
	}
	suffix := "-" + strconv.Itoa(n)
	if max := SlugMaxLength - len(suffix); len(base) > max {
		base = base[:max]
	}
	return strings.TrimRight(base, "-") + suffix
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
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrWorkspaceSlugTaken   = errors.New("workspace slug taken")
	ErrWorkspaceSlugInvalid = errors.New("workspace slug invalid")
	ErrSetupAlreadyDone     = errors.New("setup already done")
	ErrNotMember            = errors.New("not a workspace member")
)

func ValidSlugFormat(slug string) bool {
	return slugRe.MatchString(slug)
}

func (n NewWorkspace) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Slug, validation.By(slugField)),
		validation.Field(&n.Name, validation.By(workspaceNameField)),
		validation.Field(&n.Timezone, validation.By(timezoneField)),
	)
}

func (u WorkspaceUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(workspaceNameField)),
		validation.Field(&u.Timezone, validation.By(timezoneField)),
	)
}

func (s Setup) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.UserName, validation.By(nameField)),
		validation.Field(&s.Email, validation.By(emailField)),
		validation.Field(&s.Password, validation.By(passwordField)),
		validation.Field(&s.WorkspaceName, validation.By(workspaceNameField)),
		validation.Field(&s.WorkspaceSlug, validation.By(slugField)),
		validation.Field(&s.Timezone, validation.By(timezoneField)),
	)
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
