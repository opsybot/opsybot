package ntfy

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
	pkgntfy "github.com/opsybot/opsybot/internal/pkg/ntfy"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	client pkgntfy.Client
}

func New(client pkgntfy.Client) repository.Ntfy {
	return &repo{client: client}
}

func (r *repo) Publish(ctx context.Context, msg entity.NtfyMessage) (entity.NotifyResult, error) {
	published, err := r.client.Publish(ctx, msg.Server, msg.Token, pkgntfy.Message{
		Topic:    msg.Topic,
		Title:    msg.Title,
		Message:  msg.Body,
		Priority: msg.Priority,
		Click:    msg.Click,
	})
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}, nil
	}
	return entity.NotifyResult{Delivered: true, Detail: "ntfy published", MessageID: published.ID}, nil
}
