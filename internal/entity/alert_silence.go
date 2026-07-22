package entity

import (
	"errors"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SilenceState string

const (
	SilenceActive    SilenceState = "active"
	SilenceScheduled SilenceState = "scheduled"
	SilenceEnded     SilenceState = "ended"
)

type SilenceKind string

const (
	SilenceKindSilence     SilenceKind = "silence"
	SilenceKindMaintenance SilenceKind = "maintenance"
)

type SilenceCondition struct {
	Field string
	Value string
}

type Silence struct {
	ID          string
	WorkspaceID string
	Kind        SilenceKind
	Reason      string
	CreatedBy   string
	Conditions  []SilenceCondition
	StartsAt    time.Time
	EndsAt      time.Time
	CreatedAt   time.Time
}

type NewSilence struct {
	Kind       SilenceKind
	Reason     string
	Conditions []SilenceCondition
	StartsAt   time.Time
	EndsAt     time.Time
}

var SilenceScopeFields = []string{"source", "service", "label"}

var (
	ErrSilenceNotFound = errors.New("silence not found")
	ErrSilenceWindow   = errors.New("silence window invalid")
	ErrSilenceEnded    = errors.New("silence already ended")
)

func (s Silence) State(now time.Time) SilenceState {
	switch {
	case !s.EndsAt.IsZero() && !now.Before(s.EndsAt):
		return SilenceEnded
	case now.Before(s.StartsAt):
		return SilenceScheduled
	default:
		return SilenceActive
	}
}

func (s Silence) Matches(a Alert) bool {
	if len(s.Conditions) == 0 {
		return false
	}
	for _, c := range s.Conditions {
		if !c.matches(a) {
			return false
		}
	}
	return true
}

func (c SilenceCondition) matches(a Alert) bool {
	want := strings.TrimSpace(c.Value)
	switch c.Field {
	case "source":
		return strings.EqualFold(a.SourceLabel, want)
	case "service":
		return strings.EqualFold(a.ServiceName, want)
	case "label":
		key, value, found := strings.Cut(want, "=")
		if !found {
			_, present := a.Labels[strings.TrimSpace(want)]
			return present
		}
		actual, present := a.Labels[strings.TrimSpace(key)]
		return present && strings.EqualFold(actual, strings.TrimSpace(value))
	default:
		return false
	}
}

func SilenceFor(silences []Silence, a Alert, now time.Time) (Silence, bool) {
	for _, s := range silences {
		if s.State(now) == SilenceActive && s.Matches(a) {
			return s, true
		}
	}
	return Silence{}, false
}

func (n NewSilence) Validate() error {
	if n.EndsAt.IsZero() || !n.EndsAt.After(n.StartsAt) {
		return ErrSilenceWindow
	}
	return validation.ValidateStruct(&n,
		validation.Field(&n.Reason, validation.By(silenceReasonField)),
		validation.Field(&n.Conditions, validation.By(silenceConditionsField)),
	)
}
