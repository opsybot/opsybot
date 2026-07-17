package service

//go:generate go tool mockgen -source=audits.go -destination=./audits/audits_mock.go -package=audits

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Audits interface {
	List(ctx context.Context, workspaceSlug string, filter entity.AuditFilter) (entity.AuditPage, error)
}
