package http

import (
	"strings"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

func TestBuildICS(t *testing.T) {
	now := time.Date(2026, 7, 13, 8, 0, 0, 0, time.UTC)
	shifts := []entity.FeedShift{{
		StartsAt: time.Date(2026, 7, 13, 9, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC),
		UserID:   "u1",
		UserName: "Maya, Chen",
	}}
	out := string(buildICS("payments-primary", shifts, now))

	for _, want := range []string{
		"BEGIN:VCALENDAR\r\n",
		"VERSION:2.0\r\n",
		"END:VCALENDAR\r\n",
		"BEGIN:VEVENT\r\n",
		"DTSTART:20260713T090000Z\r\n",
		"DTEND:20260714T090000Z\r\n",
		"DTSTAMP:20260713T080000Z\r\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("iCal missing %q\n---\n%s", want, out)
		}
	}
	if !strings.Contains(out, "SUMMARY:On-call: Maya\\, Chen\r\n") {
		t.Errorf("comma in name not escaped:\n%s", out)
	}
	if strings.Contains(out, "\n\n") {
		t.Error("iCal contains a bare LF sequence")
	}
}

func TestBuildICSFallsBackToUserID(t *testing.T) {
	shift := entity.FeedShift{StartsAt: time.Unix(0, 0).UTC(), EndsAt: time.Unix(3600, 0).UTC(), UserID: "u9"}
	out := string(buildICS("s", []entity.FeedShift{shift}, time.Unix(0, 0).UTC()))
	if !strings.Contains(out, "SUMMARY:On-call: u9\r\n") {
		t.Errorf("expected fallback to user id:\n%s", out)
	}
}
