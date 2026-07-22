package entity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	MappingFieldTitle       = "title"
	MappingFieldDescription = "description"
	MappingFieldSeverity    = "severity"
	MappingFieldService     = "service"
	MappingFieldSource      = "source"
	MappingFieldDedupKey    = "dedup_key"
	MappingFieldStatus      = "status"
	MappingFieldStartsAt    = "starts_at"
	MappingFieldEndsAt      = "ends_at"
	MappingFieldLabels      = "labels"
)

var MappingFields = []string{
	MappingFieldTitle,
	MappingFieldDescription,
	MappingFieldSeverity,
	MappingFieldService,
	MappingFieldSource,
	MappingFieldDedupKey,
	MappingFieldStatus,
	MappingFieldStartsAt,
	MappingFieldEndsAt,
	MappingFieldLabels,
}

func LookupPath(doc any, path string) (any, bool) {
	cur := doc
	for _, seg := range splitPath(path) {
		if seg == "" {
			return nil, false
		}
		if idx, err := strconv.Atoi(seg); err == nil {
			list, ok := cur.([]any)
			if !ok || idx < 0 || idx >= len(list) {
				return nil, false
			}
			cur = list[idx]
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := obj[seg]
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

func splitPath(path string) []string {
	replaced := strings.NewReplacer("[", ".", "]", "").Replace(strings.TrimSpace(path))
	parts := strings.Split(replaced, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func PathString(doc any, path string) (string, bool) {
	v, ok := LookupPath(doc, path)
	if !ok || v == nil {
		return "", false
	}
	return scalarString(v)
}

func scalarString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case bool:
		return strconv.FormatBool(t), true
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10), true
		}
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case json.Number:
		return t.String(), true
	default:
		return "", false
	}
}

func PathLabels(doc any, path string) map[string]string {
	v, ok := LookupPath(doc, path)
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(obj))
	for k, raw := range obj {
		if s, ok := scalarString(raw); ok {
			out[k] = s
		}
	}
	return out
}

func MappingPathFor(mappings []SourceMapping, field string) string {
	for _, m := range mappings {
		if m.Field == field {
			return m.Path
		}
	}
	return ""
}

func DescribeMappingField(field string) string {
	return fmt.Sprintf("alert %s", strings.ReplaceAll(field, "_", " "))
}
