package entity

import (
	"errors"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Rotation string

const (
	RotationDaily  Rotation = "daily"
	RotationWeekly Rotation = "weekly"
	RotationCustom Rotation = "custom"
)

const (
	ScheduleSlugMaxLength     = 40
	ScheduleSlugMaxCandidates = 100
	ScheduleMaxLayers         = 10
	LayerMaxParticipants      = 50
	LayerMaxRestrictions      = 28
	LayerMinIntervalDays      = 1
	LayerMaxIntervalDays      = 30
	MinutesPerDay             = 1440
	ScheduleReasonMaxLength   = 200
	FeedTokenLength           = 18
	ScheduleHandoverLimit     = 50
	FeedPastDays              = 7
	FeedFutureDays            = 90
)

var ScheduleReservedSlugs = []string{"new", "mine", "preview"}

type Restriction struct {
	Weekday     int
	StartMinute int
	EndMinute   int
}

type Layer struct {
	ID           string
	Participants []string
	Rotation     Rotation
	IntervalDays int
	HandoverHour int
	StartsOn     time.Time
	Restrictions []Restriction
}

type Override struct {
	ID              string
	UserID          string
	StartsAt        time.Time
	EndsAt          time.Time
	Reason          string
	CreatedByUserID string
	CreatedAt       time.Time
}

type Schedule struct {
	ID          string
	WorkspaceID string
	TeamID      string
	TeamSlug    string
	Slug        string
	Timezone    string
	FeedToken   string
	Layers      []Layer
	Overrides   []Override
	Paused      bool
	Archived    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewScheduleLayer struct {
	Participants []string
	Rotation     Rotation
	IntervalDays int
	HandoverHour int
	StartsOn     time.Time
	Restrictions []Restriction
}

type NewSchedule struct {
	Slug     string
	TeamSlug string
	Timezone string
	Layers   []NewScheduleLayer
}

type ScheduleUpdate struct {
	Slug     string
	TeamSlug string
	Timezone string
	Layers   []NewScheduleLayer
}

type NewOverride struct {
	UserID   string
	StartsAt time.Time
	EndsAt   time.Time
	Reason   string
}

var (
	ErrScheduleNotFound         = errors.New("schedule not found")
	ErrScheduleSlugTaken        = errors.New("schedule slug taken")
	ErrScheduleArchived         = errors.New("schedule archived")
	ErrScheduleNotArchived      = errors.New("schedule not archived")
	ErrScheduleNotPaused        = errors.New("schedule not paused")
	ErrSchedulePaused           = errors.New("schedule paused")
	ErrScheduleTeamInvalid      = errors.New("schedule team invalid")
	ErrScheduleParticipant      = errors.New("schedule participant invalid")
	ErrScheduleOverrideConflict = errors.New("schedule override conflict")
	ErrScheduleOverrideWindow   = errors.New("schedule override window invalid")
	ErrScheduleOverrideNoChange = errors.New("schedule override no change")
)

func (r Rotation) Validate() error {
	return rotationField(r)
}

func (r Rotation) PeriodDays(intervalDays int) int {
	switch r {
	case RotationDaily:
		return 1
	case RotationWeekly:
		return 7
	default:
		if intervalDays < LayerMinIntervalDays {
			return LayerMinIntervalDays
		}
		return intervalDays
	}
}

func (n NewSchedule) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Slug, validation.By(scheduleSlugField)),
		validation.Field(&n.TeamSlug, validation.By(scheduleTeamField)),
		validation.Field(&n.Timezone, validation.By(timezoneField)),
		validation.Field(&n.Layers, validation.By(scheduleLayersField)),
	)
}

func (u ScheduleUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Slug, validation.By(scheduleSlugField)),
		validation.Field(&u.TeamSlug, validation.By(scheduleTeamField)),
		validation.Field(&u.Timezone, validation.By(timezoneField)),
		validation.Field(&u.Layers, validation.By(scheduleLayersField)),
	)
}

func (o NewOverride) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.UserID, validation.By(overrideUserField)),
		validation.Field(&o.Reason, validation.By(overrideReasonField)),
	)
}

func ScheduleSlugCandidate(base string, n int) string {
	if n <= 1 {
		return base
	}
	suffix := "-" + strconv.Itoa(n)
	if max := ScheduleSlugMaxLength - len(suffix); len(base) > max {
		base = base[:max]
	}
	return strings.TrimRight(base, "-") + suffix
}
