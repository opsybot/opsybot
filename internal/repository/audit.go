package repository

//go:generate go tool mockgen -source=audit.go -destination=./audit/audit_mock.go -package=audit

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Audit interface {
	Create(ctx context.Context, event entity.AuditEvent) error
	List(ctx context.Context, workspaceID string, filter entity.AuditFilter) ([]entity.AuditEvent, string, error)
}
