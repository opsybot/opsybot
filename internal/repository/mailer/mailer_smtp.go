package mailer

import (
	"context"

	pkgmailer "github.com/opsybot/opsybot/internal/pkg/mailer"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	client pkgmailer.Client
}

func New(client pkgmailer.Client) repository.Mailer {
	return &repo{client: client}
}

func (r *repo) SendInvite(ctx context.Context, to, inviterName, workspaceName, acceptURL string) error {
	return r.client.SendInvite(ctx, to, pkgmailer.InviteData{
		InviterName:   inviterName,
		WorkspaceName: workspaceName,
		AcceptURL:     acceptURL,
	})
}

func (r *repo) SendPasswordReset(ctx context.Context, to, resetURL string) error {
	return r.client.SendPasswordReset(ctx, to, pkgmailer.ResetData{ResetURL: resetURL})
}
