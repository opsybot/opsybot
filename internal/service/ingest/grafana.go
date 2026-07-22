package ingest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type grafanaPayload struct {
	Version           string            `json:"version"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	ExternalURL       string            `json:"externalURL"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	Alerts            []grafanaAlert    `json:"alerts"`
}

type grafanaAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
	SilenceURL   string            `json:"silenceURL"`
	DashboardURL string            `json:"dashboardURL"`
	PanelURL     string            `json:"panelURL"`
	ValueString  string            `json:"valueString"`
}

var grafanaStaleReasons = map[string]struct{}{
	"nodata":        {},
	"error":         {},
	"missingseries": {},
	"paused":        {},
	"ruledeleted":   {},
	"updated":       {},
	"keeplast":      {},
}

func parseGrafana(body []byte, src entity.AlertSource, now time.Time) ([]entity.IngestedAlert, error) {
	var payload grafanaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, failure(entity.FailureInvalidJSON, err.Error())
	}
	if len(payload.Alerts) == 0 {
		return nil, failure(entity.FailureNoAlerts, "The payload carried no alerts.")
	}

	out := make([]entity.IngestedAlert, 0, len(payload.Alerts))
	for _, a := range payload.Alerts {
		labels := mergeLabels(payload.CommonLabels, a.Labels)
		title := firstNonEmpty(a.Annotations["summary"], labels["alertname"], payload.CommonAnnotations["summary"])
		if title == "" {
			continue
		}

		startsAt, err := parseRFC3339(a.StartsAt)
		if err != nil {
			return nil, failure(entity.FailureBadTimestamp, fmt.Sprintf("startsAt %q could not be read.", a.StartsAt))
		}
		endsAt, err := parseRFC3339(a.EndsAt)
		if err != nil {
			return nil, failure(entity.FailureBadTimestamp, fmt.Sprintf("endsAt %q could not be read.", a.EndsAt))
		}

		resolved := strings.EqualFold(a.Status, "resolved")
		mode := entity.ResolveModeSource
		reason := strings.ToLower(strings.ReplaceAll(labels["grafana_state_reason"], " ", ""))
		if _, stale := grafanaStaleReasons[reason]; stale && resolved {
			mode = entity.ResolveModeTimeout
		}

		description := firstNonEmpty(a.Annotations["description"], payload.CommonAnnotations["description"], a.ValueString)

		out = append(out, entity.IngestedAlert{
			DedupKeyRaw: firstNonEmpty(a.Fingerprint, entity.CanonicalLabels(labels)),
			Title:       title,
			Description: description,
			Severity:    entity.NormalizeSeverity(labels["severity"], src.DefaultSeverity),
			SourceLabel: firstNonEmpty(labels["job"], payload.Receiver, src.Slug),
			ServiceName: firstNonEmpty(labels["service"], labels["job"]),
			Labels:      labels,
			Links:       grafanaLinks(a),
			StartedAt:   startsAt,
			EndedAt:     endsAt,
			Resolved:    resolved,
			ResolveMode: mode,
			Payload:     string(body),
		})
	}

	if len(out) == 0 {
		return nil, failure(entity.FailureMissingTitle, "No alert in the payload had a summary or alertname.")
	}
	return out, nil
}

func grafanaLinks(a grafanaAlert) []entity.AlertLink {
	links := make([]entity.AlertLink, 0, 3)
	if a.DashboardURL != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkDashboard, Label: "Dashboard", URL: a.DashboardURL})
	}
	if a.PanelURL != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkDashboard, Label: "Panel", URL: a.PanelURL})
	}
	if runbook := firstNonEmpty(a.Annotations["runbook_url"]); runbook != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkRunbook, Label: "Runbook", URL: runbook})
	}
	if a.GeneratorURL != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkSource, Label: "View in Grafana", URL: a.GeneratorURL})
	}
	return links
}
