package mailer

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
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

func (r *repo) SendPage(ctx context.Context, to string, page entity.AlertPage) error {
	return r.client.SendPage(ctx, to, pkgmailer.PageData{
		Severity:   string(page.Severity),
		Service:    page.Service,
		Title:      page.Title,
		StartedAt:  page.StartedAt.UTC().Format("2006-01-02 15:04"),
		PolicySlug: page.PolicySlug,
		Level:      page.Level,
		AlertURL:   page.AlertURL,
	})
}
