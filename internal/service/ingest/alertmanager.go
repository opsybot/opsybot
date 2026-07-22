package ingest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

const alertmanagerVersion = "4"

type alertmanagerPayload struct {
	Version           string              `json:"version"`
	GroupKey          string              `json:"groupKey"`
	TruncatedAlerts   int                 `json:"truncatedAlerts"`
	Status            string              `json:"status"`
	Receiver          string              `json:"receiver"`
	ExternalURL       string              `json:"externalURL"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	Alerts            []alertmanagerAlert `json:"alerts"`
}

type alertmanagerAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

func parseAlertmanager(body []byte, src entity.AlertSource, now time.Time) ([]entity.IngestedAlert, error) {
	var payload alertmanagerPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, failure(entity.FailureInvalidJSON, err.Error())
	}
	if payload.Version != "" && payload.Version != alertmanagerVersion {
		return nil, failure(entity.FailureUnsupportedVersion, fmt.Sprintf("Alertmanager webhook version %s is not supported.", payload.Version))
	}
	if len(payload.Alerts) == 0 {
		return nil, failure(entity.FailureNoAlerts, "The payload carried no alerts.")
	}

	repeat := strings.Contains(strings.ToLower(payload.CommonAnnotations["notification_reason"]), "repeat interval")

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

		out = append(out, entity.IngestedAlert{
			DedupKeyRaw: firstNonEmpty(a.Fingerprint, entity.CanonicalLabels(labels)),
			Title:       title,
			Description: firstNonEmpty(a.Annotations["description"], payload.CommonAnnotations["description"]),
			Severity:    entity.NormalizeSeverity(labels["severity"], src.DefaultSeverity),
			SourceLabel: firstNonEmpty(labels["job"], payload.Receiver, src.Slug),
			ServiceName: firstNonEmpty(labels["service"], labels["job"], labels["namespace"]),
			Labels:      labels,
			Links:       alertmanagerLinks(a, payload),
			StartedAt:   startsAt,
			EndedAt:     endsAt,
			Resolved:    resolved,
			ResolveMode: entity.ResolveModeSource,
			Repeat:      repeat,
			Payload:     string(body),
		})
	}

	if len(out) == 0 {
		return nil, failure(entity.FailureMissingTitle, "No alert in the payload had a summary or alertname.")
	}
	return out, nil
}

func alertmanagerLinks(a alertmanagerAlert, payload alertmanagerPayload) []entity.AlertLink {
	links := make([]entity.AlertLink, 0, 3)
	if a.GeneratorURL != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkSource, Label: "View in Prometheus", URL: a.GeneratorURL})
	}
	if runbook := firstNonEmpty(a.Annotations["runbook_url"], payload.CommonAnnotations["runbook_url"]); runbook != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkRunbook, Label: "Runbook", URL: runbook})
	}
	if dashboard := firstNonEmpty(a.Annotations["dashboard_url"], payload.CommonAnnotations["dashboard_url"]); dashboard != "" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkDashboard, Label: "Dashboard", URL: dashboard})
	}
	return links
}
