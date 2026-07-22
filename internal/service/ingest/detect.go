package ingest

import (
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

const zeroTimestamp = "0001-01-01T00:00:00Z"

func failure(reason entity.IngestFailureReason, detail string) error {
	return entity.ParseFailure(reason, detail)
}

func parseFor(format entity.SourceFormat, body []byte, src entity.AlertSource, now time.Time) ([]entity.IngestedAlert, error) {
	switch format {
	case entity.SourceFormatAlertmanager:
		return parseAlertmanager(body, src, now)
	case entity.SourceFormatGrafana:
		return parseGrafana(body, src, now)
	case entity.SourceFormatKuma:
		return parseKuma(body, src, now)
	case entity.SourceFormatHeartbeat, entity.SourceFormatGeneric:
		return parseGeneric(body, src, now)
	default:
		return nil, failure(entity.FailureUnsupportedFormat, fmt.Sprintf("Format %q is not supported.", format))
	}
}

func parseRFC3339(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == zeroTimestamp {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.Year() <= 1 {
		return time.Time{}, nil
	}
	return parsed.UTC(), nil
}

func mergeLabels(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		if !strings.HasPrefix(k, "__") {
			out[k] = v
		}
	}
	for k, v := range extra {
		if !strings.HasPrefix(k, "__") {
			out[k] = v
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
