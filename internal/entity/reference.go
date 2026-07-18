package entity

import "errors"

type MemberReference struct {
	ID     string
	Kind   string
	Icon   string
	Label  string
	Detail string
}

const (
	ReferenceKindSchedule = "schedule"
	ReferenceKindPolicy   = "policy"
	ReferenceKindFollowup = "followup"
)

var ErrReferenceUnknown = errors.New("reference unknown")
