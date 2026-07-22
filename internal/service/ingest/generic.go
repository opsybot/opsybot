package ingest

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

func parseGeneric(body []byte, src entity.AlertSource, now time.Time) ([]entity.IngestedAlert, error) {
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, failure(entity.FailureInvalidJSON, err.Error())
	}

	root, ok := doc.(map[string]any)
	if !ok {
		return nil, failure(entity.FailureInvalidJSON, "The payload must be a JSON object.")
	}

	items := []map[string]any{root}
	if raw, found := root["alerts"]; found {
		list, isList := raw.([]any)
		if !isList {
			return nil, failure(entity.FailureInvalidJSON, "The alerts field must be an array.")
		}
		if len(list) == 0 {
			return nil, failure(entity.FailureNoAlerts, "The payload carried no alerts.")
		}
		items = items[:0]
		for _, entry := range list {
			obj, isObj := entry.(map[string]any)
			if !isObj {
				return nil, failure(entity.FailureInvalidJSON, "Each alert must be a JSON object.")
			}
			items = append(items, obj)
		}
	}

	out := make([]entity.IngestedAlert, 0, len(items))
	for _, item := range items {
		alert, err := genericAlert(item, root, src, now, body)
		if err != nil {
			return nil, err
		}
		out = append(out, alert)
	}
	return out, nil
}

func genericAlert(item, root map[string]any, src entity.AlertSource, now time.Time, body []byte) (entity.IngestedAlert, error) {
	pick := func(field, fallback string) string {
		if path := entity.MappingPathFor(src.Mapping, field); path != "" {
			if v, ok := entity.PathString(item, path); ok {
				return v
			}
			if v, ok := entity.PathString(root, path); ok {
				return v
			}
			return ""
		}
		if v, ok := entity.PathString(item, fallback); ok {
			return v
		}
		return ""
	}

	title := pick(entity.MappingFieldTitle, "title")
	if title == "" {
		title = firstNonEmpty(genericString(item, "summary"), genericString(item, "message"), genericString(item, "name"))
	}
	if title == "" {
		return entity.IngestedAlert{}, failure(entity.FailureMissingTitle, "The payload has no title, summary, or message.")
	}

	labelsPath := entity.MappingPathFor(src.Mapping, entity.MappingFieldLabels)
	if labelsPath == "" {
		labelsPath = "labels"
	}
	labels := entity.PathLabels(item, labelsPath)
	if labels == nil {
		labels = entity.PathLabels(root, labelsPath)
	}
	if labels == nil {
		labels = map[string]string{}
	}

	startedAt, err := genericTime(pick(entity.MappingFieldStartsAt, "starts_at"), now)
	if err != nil {
		return entity.IngestedAlert{}, failure(entity.FailureBadTimestamp, "starts_at could not be read.")
	}
	endedAt, err := genericTime(pick(entity.MappingFieldEndsAt, "ends_at"), time.Time{})
	if err != nil {
		return entity.IngestedAlert{}, failure(entity.FailureBadTimestamp, "ends_at could not be read.")
	}

	status := strings.ToLower(strings.TrimSpace(pick(entity.MappingFieldStatus, "status")))
	resolved := status == "resolved" || status == "ok" || status == "up" || status == "success"

	return entity.IngestedAlert{
		DedupKeyRaw: pick(entity.MappingFieldDedupKey, "dedup_key"),
		Title:       title,
		Description: pick(entity.MappingFieldDescription, "description"),
		Severity:    entity.NormalizeSeverity(pick(entity.MappingFieldSeverity, "severity"), src.DefaultSeverity),
		SourceLabel: firstNonEmpty(pick(entity.MappingFieldSource, "source"), src.Slug),
		ServiceName: pick(entity.MappingFieldService, "service"),
		Labels:      labels,
		StartedAt:   startedAt,
		EndedAt:     endedAt,
		Resolved:    resolved,
		ResolveMode: entity.ResolveModeSource,
		Payload:     string(body),
	}, nil
}

func genericString(doc map[string]any, key string) string {
	v, ok := entity.PathString(doc, key)
	if !ok {
		return ""
	}
	return v
}

func genericTime(raw string, fallback time.Time) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	parsed, err := parseRFC3339(raw)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.IsZero() {
		return fallback, nil
	}
	return parsed, nil
}
