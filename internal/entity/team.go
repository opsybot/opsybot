package entity

import (
	"errors"
	"strconv"
	"strings"
	"time"
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
	ErrTeamNotFound       = errors.New("team not found")
	ErrTeamSlugTaken      = errors.New("team slug taken")
	ErrTeamNameInvalid    = errors.New("team name invalid")
	ErrTeamArchived       = errors.New("team archived")
	ErrTeamNotArchived    = errors.New("team not archived")
	ErrTeamTooManyMembers = errors.New("team too many members")
	ErrTeamMemberInvalid  = errors.New("team member invalid")
)

func ValidateTeamName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > TeamNameMaxLength {
		return ErrTeamNameInvalid
	}
	return nil
}

func (n NewTeam) Validate() error {
	if err := ValidateTeamName(n.Name); err != nil {
		return err
	}
	if len(n.MemberIDs) > TeamMaxMembers {
		return ErrTeamTooManyMembers
	}
	return nil
}

func (u TeamUpdate) Validate() error {
	if err := ValidateTeamName(u.Name); err != nil {
		return err
	}
	if len(u.MemberIDs) > TeamMaxMembers {
		return ErrTeamTooManyMembers
	}
	return nil
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
