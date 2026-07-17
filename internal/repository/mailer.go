package repository

//go:generate go tool mockgen -source=mailer.go -destination=./mailer/mailer_mock.go -package=mailer

import "context"

type Mailer interface {
	SendInvite(ctx context.Context, to, inviterName, workspaceName, acceptURL string) error
	SendPasswordReset(ctx context.Context, to, resetURL string) error
}
