package entity

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SourceFormat string

const (
	SourceFormatAlertmanager SourceFormat = "alertmanager"
	SourceFormatGrafana      SourceFormat = "grafana"
	SourceFormatKuma         SourceFormat = "kuma"
	SourceFormatHeartbeat    SourceFormat = "heartbeat"
	SourceFormatGeneric      SourceFormat = "generic"
)

type SourceHealth string

const (
	SourceHealthHealthy SourceHealth = "healthy"
	SourceHealthStale   SourceHealth = "stale"
	SourceHealthFailing SourceHealth = "failing"
	SourceHealthPaused  SourceHealth = "paused"
)

const (
	SourceSlugMaxLength    = 40
	SourceNameMaxLength    = 60
	SourceMaxMappings      = 32
	SourceMappingPathMax   = 200
	IngestTokenLength      = 24
	SigningSecretByteLen   = 24
	SourceStaleAfter       = 72 * time.Hour
	SourceFailureWindow    = 24 * time.Hour
	SourceSecretGrace      = 24 * time.Hour
	SourceSignaturePrefix  = "sha256="
	SourceSignatureHeader  = "X-Opsy-Signature"
	SourceMaxAutoResolve   = 30 * 24 * time.Hour
	SourceEventSampleBytes = 4096
)

var SourceReservedSlugs = []string{"new", "routing", "formats"}

type SourceMapping struct {
	Field    string
	Path     string
	Position int
}

type AlertSource struct {
	ID                    string
	WorkspaceID           string
	Slug                  string
	Name                  string
	Format                SourceFormat
	IngestToken           string
	SigningSecret         string
	SigningSecretPrevious string
	SecretRotatedAt       time.Time
	RequireSignature      bool
	DefaultSeverity       AlertSeverity
	AutoResolveAfter      time.Duration
	Mapping               []SourceMapping
	LastEventAt           time.Time
	FailureCount          int
	Paused                bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type NewAlertSource struct {
	Slug             string
	Name             string
	Format           SourceFormat
	DefaultSeverity  AlertSeverity
	RequireSignature bool
	AutoResolveAfter time.Duration
}

type AlertSourceUpdate struct {
	Name             string
	DefaultSeverity  AlertSeverity
	RequireSignature bool
	AutoResolveAfter time.Duration
}

var (
	ErrAlertSourceNotFound      = errors.New("alert source not found")
	ErrAlertSourceSlugTaken     = errors.New("alert source slug taken")
	ErrAlertSourcePaused        = errors.New("alert source paused")
	ErrAlertSourceMappingEmpty  = errors.New("alert source mapping empty")
	ErrAlertSourceSignature     = errors.New("alert source signature invalid")
	ErrAlertSourceFormatInvalid = errors.New("alert source format invalid")
)

func (f SourceFormat) Validate() error {
	return sourceFormatField(f)
}

func (s AlertSource) Health(now time.Time) SourceHealth {
	switch {
	case s.Paused:
		return SourceHealthPaused
	case s.FailureCount > 0:
		return SourceHealthFailing
	case s.LastEventAt.IsZero() || now.Sub(s.LastEventAt) > SourceStaleAfter:
		return SourceHealthStale
	default:
		return SourceHealthHealthy
	}
}

func (s AlertSource) SecretsInGrace(now time.Time) []string {
	out := make([]string, 0, 2)
	if s.SigningSecret != "" {
		out = append(out, s.SigningSecret)
	}
	if s.SigningSecretPrevious != "" && !s.SecretRotatedAt.IsZero() && now.Sub(s.SecretRotatedAt) < SourceSecretGrace {
		out = append(out, s.SigningSecretPrevious)
	}
	return out
}

func (n NewAlertSource) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Slug, validation.By(sourceSlugField)),
		validation.Field(&n.Name, validation.By(sourceNameField)),
		validation.Field(&n.Format, validation.By(sourceFormatField)),
		validation.Field(&n.DefaultSeverity, validation.By(alertSeverityField)),
		validation.Field(&n.AutoResolveAfter, validation.By(autoResolveField)),
	)
}

func (u AlertSourceUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(sourceNameField)),
		validation.Field(&u.DefaultSeverity, validation.By(alertSeverityField)),
		validation.Field(&u.AutoResolveAfter, validation.By(autoResolveField)),
	)
}

func ValidateSourceMappings(format SourceFormat, mappings []SourceMapping) error {
	if format != SourceFormatGeneric {
		return nil
	}
	if len(mappings) == 0 {
		return ErrAlertSourceMappingEmpty
	}
	return validation.Validate(mappings, validation.By(sourceMappingsField))
}
