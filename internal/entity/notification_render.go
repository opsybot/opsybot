package entity

import (
	"strconv"
	"strings"
	"unicode"
)

func (p AlertPage) Subject() string {
	sev := severityLabel(p.Severity)
	if p.Service != "" {
		return sev + " · " + p.Service + ": " + p.Title
	}
	return sev + ": " + p.Title
}

func (p AlertPage) BodyLines() []string {
	lines := []string{severityLabel(p.Severity) + " · " + p.Service, p.Title}
	lines = append(lines, "Started "+p.StartedAt.UTC().Format("2006-01-02 15:04")+" UTC")
	if p.PolicySlug != "" {
		lines = append(lines, "Paged by the "+p.PolicySlug+" policy, level "+strconv.Itoa(p.Level))
	}
	if p.AlertURL != "" {
		lines = append(lines, "Acknowledge or resolve: "+p.AlertURL)
	}
	return lines
}

func (p AlertPage) PlainText() string {
	return strings.Join(p.BodyLines(), "\n")
}

func severityLabel(sev AlertSeverity) string {
	s := string(sev)
	if s == "" {
		return "Alert"
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
