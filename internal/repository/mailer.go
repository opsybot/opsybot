package repository

//go:generate go tool mockgen -source=mailer.go -destination=./mailer/mailer_mock.go -package=mailer

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Mailer interface {
	SendInvite(ctx context.Context, to, inviterName, workspaceName, acceptURL string) error
	SendPasswordReset(ctx context.Context, to, resetURL string) error
	SendPage(ctx context.Context, to string, page entity.AlertPage) error
}
