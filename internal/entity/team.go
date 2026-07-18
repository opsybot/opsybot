package entity

import (
	"errors"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	TeamNameMaxLength     = 60
	TeamMaxMembers        = 50
	TeamSlugMaxLength     = 40
	TeamSlugMaxCandidates = 100
)

var TeamReservedSlugs = []string{"teams", "new", "keys", "audit", "settings", "config"}

type Team struct {
	ID          string
	WorkspaceID string
	Slug        string
	Name        string
	MemberIDs   []string
	Archived    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewTeam struct {
	Name      string
	MemberIDs []string
}

type TeamUpdate struct {
	Name      string
	MemberIDs []string
}

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamSlugTaken     = errors.New("team slug taken")
	ErrTeamArchived      = errors.New("team archived")
	ErrTeamNotArchived   = errors.New("team not archived")
	ErrTeamMemberInvalid = errors.New("team member invalid")
)

func (n NewTeam) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Name, validation.By(teamNameField)),
		validation.Field(&n.MemberIDs, validation.By(teamMembersField)),
	)
}

func (u TeamUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(teamNameField)),
		validation.Field(&u.MemberIDs, validation.By(teamMembersField)),
	)
}

func TeamSlugCandidate(base string, n int) string {
	if n <= 1 {
		return base
	}
	suffix := "-" + strconv.Itoa(n)
	if max := TeamSlugMaxLength - len(suffix); len(base) > max {
		base = base[:max]
	}
	return strings.TrimRight(base, "-") + suffix
}
