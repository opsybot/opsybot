package entity

import (
	"slices"
	"time"
)

func EscalationPriorityLane(severity AlertSeverity) string {
	if severity == SeverityWarning {
		return EscalationLaneLow
	}
	return EscalationLaneHigh
}

func (h HoursWindow) Contains(at time.Time) bool {
	loc, err := time.LoadLocation(h.Timezone)
	if err != nil {
		loc = time.UTC
	}
	local := at.In(loc)
	if !slices.Contains(h.Days, int(local.Weekday())) {
		return false
	}
	minute := local.Hour()*60 + local.Minute()
	if h.StartMinute <= h.EndMinute {
		return minute >= h.StartMinute && minute < h.EndMinute
	}
	return minute >= h.StartMinute || minute < h.EndMinute
}

func DefaultHoursWindow() HoursWindow {
	return HoursWindow{
		Days:        []int{1, 2, 3, 4, 5},
		StartMinute: EscalationDefaultStartMin,
		EndMinute:   EscalationDefaultEndMin,
		Timezone:    "UTC",
	}
}

func (b EscalationBranch) LaneKeyFor(severity AlertSeverity, at time.Time) string {
	switch b.On {
	case EscalationBranchHours:
		if b.Hours.Contains(at) {
			return EscalationLaneBusiness
		}
		return EscalationLaneOff
	default:
		return EscalationPriorityLane(severity)
	}
}

func FlattenEscalation(nodes []EscalationNode, severity AlertSeverity, at time.Time) ([]EscalationLevel, map[string]string) {
	choices := map[string]string{}
	levels := flattenNodes(nodes, severity, at, choices)
	return levels, choices
}

func flattenNodes(nodes []EscalationNode, severity AlertSeverity, at time.Time, choices map[string]string) []EscalationLevel {
	var out []EscalationLevel
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			out = append(out, *node.Level)
		case node.Branch != nil:
			key := node.Branch.LaneKeyFor(severity, at)
			choices[node.Branch.ID] = key
			for _, lane := range node.Branch.Lanes {
				if lane.Key == key {
					out = append(out, flattenNodes(lane.Nodes, severity, at, choices)...)
				}
			}
		}
	}
	return out
}

func (l EscalationLevel) PickTargets(valid []EscalationTarget, position int) []EscalationTarget {
	if len(valid) == 0 {
		return nil
	}
	if l.Mode == NotifyModeRoundRobin {
		return []EscalationTarget{valid[position%len(valid)]}
	}
	return valid
}

type EscalationTickKind string

const (
	EscalationTickNotify  EscalationTickKind = "notify"
	EscalationTickRepeat  EscalationTickKind = "repeat"
	EscalationTickExhaust EscalationTickKind = "exhaust"
	EscalationTickResume  EscalationTickKind = "resume"
)

type EscalationTick struct {
	Kind  EscalationTickKind
	Level EscalationLevel
}

func (r EscalationRun) NextTick(now time.Time) (EscalationTick, bool) {
	switch r.State {
	case EscalationAcked:
		if !r.AckExpiresAt.IsZero() && !now.Before(r.AckExpiresAt) {
			return EscalationTick{Kind: EscalationTickResume}, true
		}
		return EscalationTick{}, false
	case EscalationRunning:
		if now.Before(r.NextAt) {
			return EscalationTick{}, false
		}
		if r.StepIndex < len(r.Path) {
			return EscalationTick{Kind: EscalationTickNotify, Level: r.Path[r.StepIndex]}, true
		}
		if r.Cycle < r.Snapshot.Repeat {
			return EscalationTick{Kind: EscalationTickRepeat}, true
		}
		return EscalationTick{Kind: EscalationTickExhaust}, true
	default:
		return EscalationTick{}, false
	}
}

func StartEscalationRun(alert Alert, policy EscalationPolicy, now time.Time) EscalationRun {
	path, choices := FlattenEscalation(policy.Nodes, alert.Severity, now)
	return EscalationRun{
		WorkspaceID: alert.WorkspaceID,
		AlertID:     alert.ID,
		PolicyID:    policy.ID,
		PolicySlug:  policy.Slug,
		State:       EscalationRunning,
		Snapshot: EscalationSnapshot{
			Repeat:     policy.Repeat,
			AckTimeout: policy.AckTimeout,
			Nodes:      policy.Nodes,
		},
		Path:        path,
		LaneChoices: choices,
		NextAt:      now,
		StartedAt:   now,
	}
}

func (r EscalationRun) Repeated(severity AlertSeverity, now time.Time) EscalationRun {
	path, choices := FlattenEscalation(r.Snapshot.Nodes, severity, now)
	out := r
	out.Cycle++
	out.StepIndex = 0
	out.Path = path
	out.LaneChoices = choices
	out.NextAt = now
	return out
}
