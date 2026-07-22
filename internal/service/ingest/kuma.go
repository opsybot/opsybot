package ingest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

const kumaTimeLayout = "2006-01-02 15:04:05.000"

const (
	kumaStatusDown        = 0
	kumaStatusUp          = 1
	kumaStatusMaintenance = 3
)

type kumaPayload struct {
	Heartbeat *kumaHeartbeat `json:"heartbeat"`
	Monitor   *kumaMonitor   `json:"monitor"`
	Msg       string         `json:"msg"`
}

type kumaHeartbeat struct {
	Status    int    `json:"status"`
	Time      string `json:"time"`
	Msg       string `json:"msg"`
	Important bool   `json:"important"`
}

type kumaMonitor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

func parseKuma(body []byte, src entity.AlertSource, now time.Time) ([]entity.IngestedAlert, error) {
	var payload kumaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, failure(entity.FailureInvalidJSON, err.Error())
	}

	if payload.Monitor == nil || payload.Heartbeat == nil {
		title := firstNonEmpty(payload.Msg, "Uptime Kuma notification")
		return []entity.IngestedAlert{{
			DedupKeyRaw: "kuma/notice/" + title,
			Title:       title,
			Description: payload.Msg,
			Severity:    entity.NormalizeSeverity("", src.DefaultSeverity),
			SourceLabel: src.Slug,
			Labels:      map[string]string{},
			StartedAt:   now,
			Payload:     string(body),
		}}, nil
	}

	at, err := parseKumaTime(payload.Heartbeat.Time, now)
	if err != nil {
		return nil, failure(entity.FailureBadTimestamp, fmt.Sprintf("heartbeat time %q could not be read.", payload.Heartbeat.Time))
	}

	name := firstNonEmpty(payload.Monitor.Name, "Uptime Kuma monitor")
	resolved := payload.Heartbeat.Status == kumaStatusUp || payload.Heartbeat.Status == kumaStatusMaintenance

	title := name + " is down"
	severity := entity.NormalizeSeverity("critical", src.DefaultSeverity)
	if resolved {
		title = name + " recovered"
		severity = entity.NormalizeSeverity("", src.DefaultSeverity)
	}

	labels := map[string]string{"monitor": name}
	if payload.Monitor.Type != "" {
		labels["type"] = payload.Monitor.Type
	}

	links := make([]entity.AlertLink, 0, 1)
	if payload.Monitor.URL != "" && payload.Monitor.URL != "https://" {
		links = append(links, entity.AlertLink{Kind: entity.AlertLinkSource, Label: "Monitored URL", URL: payload.Monitor.URL})
	}

	return []entity.IngestedAlert{{
		DedupKeyRaw: fmt.Sprintf("kuma/%d", payload.Monitor.ID),
		Title:       title,
		Description: payload.Heartbeat.Msg,
		Severity:    severity,
		SourceLabel: src.Slug,
		ServiceName: name,
		Labels:      labels,
		Links:       links,
		StartedAt:   at,
		EndedAt:     resolvedTime(resolved, at),
		Resolved:    resolved,
		ResolveMode: entity.ResolveModeSource,
		Payload:     string(body),
	}}, nil
}

func parseKumaTime(raw string, now time.Time) (time.Time, error) {
	if raw == "" {
		return now, nil
	}
	if parsed, err := time.ParseInLocation(kumaTimeLayout, raw, time.UTC); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.UTC); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func resolvedTime(resolved bool, at time.Time) time.Time {
	if resolved {
		return at
	}
	return time.Time{}
}
