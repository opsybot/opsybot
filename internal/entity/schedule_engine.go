package entity

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const onCallHorizon = 14 * 24 * time.Hour

type Coverage struct {
	UserID   string
	Via      string
	Override bool
}

type Segment struct {
	StartsAt time.Time
	EndsAt   time.Time
	UserID   string
	Via      string
	Override bool
}

type Handover struct {
	At         time.Time
	FromUserID string
	ToUserID   string
}

type Shift struct {
	StartsAt     time.Time
	EndsAt       time.Time
	ScheduleID   string
	ScheduleSlug string
}

type FeedShift struct {
	StartsAt time.Time
	EndsAt   time.Time
	UserID   string
	UserName string
}

type LayerCoverage struct {
	Index    int
	Via      string
	Segments []Segment
}

type ScheduleCalendar struct {
	Effective []Segment
	Layers    []LayerCoverage
	Gaps      []Segment
	Handovers []Handover
}

func (s Schedule) location() *time.Location {
	if loc, err := time.LoadLocation(s.Timezone); err == nil {
		return loc
	}
	return time.UTC
}

func layerVia(total, index int) string {
	return fmt.Sprintf("layer %d", total-index)
}

func layerBoundary(layer Layer, loc *time.Location, k int) time.Time {
	y, m, d := layer.StartsOn.UTC().Date()
	period := layer.Rotation.PeriodDays(layer.IntervalDays)
	return time.Date(y, m, d+k*period, layer.HandoverHour, 0, 0, 0, loc)
}

func rotationIndexAt(layer Layer, loc *time.Location, at time.Time) int {
	period := layer.Rotation.PeriodDays(layer.IntervalDays)
	anchor := layerBoundary(layer, loc, 0)
	k := int(math.Floor(at.Sub(anchor).Hours() / float64(period*24)))
	for layerBoundary(layer, loc, k).After(at) {
		k--
	}
	for !layerBoundary(layer, loc, k+1).After(at) {
		k++
	}
	return k
}

func rotationPersonAt(layer Layer, loc *time.Location, at time.Time) string {
	n := len(layer.Participants)
	if n == 0 {
		return ""
	}
	k := rotationIndexAt(layer, loc, at)
	return layer.Participants[((k%n)+n)%n]
}

func layerOnDutyAt(layer Layer, loc *time.Location, at time.Time) bool {
	if len(layer.Restrictions) == 0 {
		return true
	}
	local := at.In(loc)
	weekday := int(local.Weekday())
	minute := local.Hour()*60 + local.Minute()
	for _, r := range layer.Restrictions {
		if r.Weekday == weekday && minute >= r.StartMinute && minute < r.EndMinute {
			return true
		}
	}
	return false
}

func (s Schedule) layerCoverageAt(loc *time.Location, index int, at time.Time) (Coverage, bool) {
	layer := s.Layers[index]
	if !layerOnDutyAt(layer, loc, at) {
		return Coverage{}, false
	}
	person := rotationPersonAt(layer, loc, at)
	if person == "" {
		return Coverage{}, false
	}
	return Coverage{UserID: person, Via: layerVia(len(s.Layers), index)}, true
}

func (s Schedule) coverageAt(loc *time.Location, at time.Time) Coverage {
	for i := len(s.Overrides) - 1; i >= 0; i-- {
		o := s.Overrides[i]
		if !at.Before(o.StartsAt) && at.Before(o.EndsAt) {
			return Coverage{UserID: o.UserID, Via: "override", Override: true}
		}
	}
	for i := range s.Layers {
		if cover, ok := s.layerCoverageAt(loc, i, at); ok {
			return cover
		}
	}
	return Coverage{}
}

func (s Schedule) boundaries(loc *time.Location, from, to time.Time, layerIndex int) []time.Time {
	set := map[int64]time.Time{from.UnixNano(): from, to.UnixNano(): to}
	add := func(t time.Time) {
		if t.After(from) && t.Before(to) {
			set[t.UnixNano()] = t
		}
	}

	for _, i := range s.layerRange(layerIndex) {
		layer := s.Layers[i]
		kFrom := rotationIndexAt(layer, loc, from)
		kTo := rotationIndexAt(layer, loc, to)
		for k := kFrom; k <= kTo+1; k++ {
			add(layerBoundary(layer, loc, k).UTC())
		}
		if len(layer.Restrictions) > 0 {
			addRestrictionEdges(loc, layer, from, to, add)
		}
	}

	if layerIndex < 0 {
		for _, o := range s.Overrides {
			add(o.StartsAt)
			add(o.EndsAt)
		}
	}

	out := make([]time.Time, 0, len(set))
	for _, t := range set {
		out = append(out, t)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Before(out[b]) })
	return out
}

func addRestrictionEdges(loc *time.Location, layer Layer, from, to time.Time, add func(time.Time)) {
	local := from.In(loc)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	for !day.After(to) {
		weekday := int(day.Weekday())
		for _, r := range layer.Restrictions {
			if r.Weekday != weekday {
				continue
			}
			add(time.Date(day.Year(), day.Month(), day.Day(), 0, r.StartMinute, 0, 0, loc).UTC())
			add(time.Date(day.Year(), day.Month(), day.Day(), 0, r.EndMinute, 0, 0, loc).UTC())
		}
		day = day.AddDate(0, 0, 1)
	}
}

func (s Schedule) layerRange(layerIndex int) []int {
	if layerIndex >= 0 {
		return []int{layerIndex}
	}
	out := make([]int, len(s.Layers))
	for i := range s.Layers {
		out[i] = i
	}
	return out
}

func (s Schedule) Segments(from, to time.Time, layerIndex int) []Segment {
	if !from.Before(to) {
		return nil
	}
	loc := s.location()
	marks := s.boundaries(loc, from, to, layerIndex)
	out := make([]Segment, 0, len(marks))
	for i := 0; i < len(marks)-1; i++ {
		start, end := marks[i], marks[i+1]
		mid := start.Add(end.Sub(start) / 2)

		var cover Coverage
		if layerIndex < 0 {
			cover = s.coverageAt(loc, mid)
		} else {
			cover, _ = s.layerCoverageAt(loc, layerIndex, mid)
		}

		if n := len(out); n > 0 && out[n-1].UserID == cover.UserID && out[n-1].Override == cover.Override {
			out[n-1].EndsAt = end
			continue
		}
		out = append(out, Segment{StartsAt: start, EndsAt: end, UserID: cover.UserID, Via: cover.Via, Override: cover.Override})
	}
	return out
}

func (s Schedule) Gaps(from, to time.Time) []Segment {
	if s.Paused {
		return nil
	}
	var out []Segment
	for _, seg := range s.Segments(from, to, -1) {
		if seg.UserID == "" {
			out = append(out, seg)
		}
	}
	return out
}

func handoversFrom(runs []Segment, limit int, paused bool) []Handover {
	if paused {
		return nil
	}
	var out []Handover
	for i := 1; i < len(runs) && len(out) < limit; i++ {
		before, after := runs[i-1], runs[i]
		if before.UserID == "" || after.UserID == "" || before.UserID == after.UserID {
			continue
		}
		out = append(out, Handover{At: after.StartsAt, FromUserID: before.UserID, ToUserID: after.UserID})
	}
	return out
}

func (s Schedule) HandoverList(from, to time.Time, limit int) []Handover {
	if s.Paused {
		return nil
	}
	return handoversFrom(s.Segments(from, to, -1), limit, false)
}

func (s Schedule) Calendar(from, to time.Time, handoverLimit int) ScheduleCalendar {
	cal := ScheduleCalendar{Effective: s.Segments(from, to, -1)}
	if !s.Paused {
		for _, seg := range cal.Effective {
			if seg.UserID == "" {
				cal.Gaps = append(cal.Gaps, seg)
			}
		}
	}
	cal.Handovers = handoversFrom(cal.Effective, handoverLimit, s.Paused)
	for i := range s.Layers {
		cal.Layers = append(cal.Layers, LayerCoverage{Index: i, Via: layerVia(len(s.Layers), i), Segments: s.Segments(from, to, i)})
	}
	return cal
}

func (s Schedule) OnCallAt(at time.Time) Coverage {
	if s.Paused {
		return Coverage{}
	}
	return s.coverageAt(s.location(), at)
}

func (s Schedule) OnCallSegment(at time.Time) (Segment, bool) {
	if s.Paused {
		return Segment{}, false
	}
	runs := s.Segments(at, at.Add(onCallHorizon), -1)
	if len(runs) == 0 {
		return Segment{}, false
	}
	return runs[0], true
}

func (s Schedule) OverrideConflicts(from, to time.Time) bool {
	for _, o := range s.Overrides {
		if from.Before(o.EndsAt) && o.StartsAt.Before(to) {
			return true
		}
	}
	return false
}

func (s Schedule) OverrideRedundant(userID string, from, to time.Time) bool {
	segs := s.Segments(from, to, -1)
	if len(segs) == 0 {
		return false
	}
	for _, seg := range segs {
		if seg.UserID != userID {
			return false
		}
	}
	return true
}

func Shifts(schedules []Schedule, userID string, from, to time.Time) []Shift {
	horizon := to.Add(7 * 24 * time.Hour)
	var out []Shift
	for _, s := range schedules {
		if s.Paused {
			continue
		}
		for _, seg := range s.Segments(from, horizon, -1) {
			if seg.UserID != userID || !seg.StartsAt.Before(to) {
				continue
			}
			out = append(out, Shift{StartsAt: seg.StartsAt, EndsAt: seg.EndsAt, ScheduleID: s.ID, ScheduleSlug: s.Slug})
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].StartsAt.Before(out[b].StartsAt) })
	return out
}
