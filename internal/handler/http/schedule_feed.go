package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

const icsTimeLayout = "20060102T150405Z"

type feedRoutes struct {
	schedules service.Schedules
}

func (h *feedRoutes) serve(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSuffix(chi.URLParam(r, "token"), ".ics")
	sched, shifts, err := h.schedules.Feed(r.Context(), token)
	if err != nil {
		if errors.Is(err, entity.ErrScheduleNotFound) {
			http.NotFound(w, r)
			return
		}
		logger.From(r.Context()).WarnContext(r.Context(), "schedule feed failed", "error", err)
		http.Error(w, "feed unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = w.Write(buildICS(sched.Slug, shifts, time.Now()))
}

func buildICS(slug string, shifts []entity.FeedShift, now time.Time) []byte {
	var b strings.Builder
	stamp := now.UTC().Format(icsTimeLayout)

	foldLine(&b, "BEGIN:VCALENDAR")
	foldLine(&b, "VERSION:2.0")
	foldLine(&b, "PRODID:-//Opsybot//On-call//EN")
	foldLine(&b, "CALSCALE:GREGORIAN")
	foldLine(&b, "METHOD:PUBLISH")
	foldLine(&b, "X-WR-CALNAME:"+icsEscape(slug+" on-call"))
	for _, sh := range shifts {
		name := sh.UserName
		if name == "" {
			name = sh.UserID
		}
		foldLine(&b, "BEGIN:VEVENT")
		foldLine(&b, "UID:"+fmt.Sprintf("%s-%d-%s@opsybot", slug, sh.StartsAt.UTC().Unix(), sh.UserID))
		foldLine(&b, "DTSTAMP:"+stamp)
		foldLine(&b, "DTSTART:"+sh.StartsAt.UTC().Format(icsTimeLayout))
		foldLine(&b, "DTEND:"+sh.EndsAt.UTC().Format(icsTimeLayout))
		foldLine(&b, "SUMMARY:"+icsEscape("On-call: "+name))
		foldLine(&b, "END:VEVENT")
	}
	foldLine(&b, "END:VCALENDAR")
	return []byte(b.String())
}

func icsEscape(s string) string {
	return strings.NewReplacer("\\", "\\\\", ";", "\\;", ",", "\\,", "\n", "\\n").Replace(s)
}

func foldLine(b *strings.Builder, s string) {
	const limit = 74
	for len(s) > limit {
		cut := limit
		for cut > 0 && s[cut]&0xC0 == 0x80 {
			cut--
		}
		b.WriteString(s[:cut])
		b.WriteString("\r\n ")
		s = s[cut:]
	}
	b.WriteString(s)
	b.WriteString("\r\n")
}
