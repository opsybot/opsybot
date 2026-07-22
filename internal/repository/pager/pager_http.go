package pager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/webhook"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	client webhook.Client
}

func New(client webhook.Client) repository.Pager {
	return &repo{client: client}
}

func (r *repo) Deliver(ctx context.Context, hook entity.EscalationWebhook, payload []byte) (entity.NotifyResult, error) {
	signature := ""
	if hook.Secret != "" {
		signature = entity.SignBody(hook.Secret, payload)
	}
	status, err := r.client.Post(ctx, hook.URL, signature, payload)
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return entity.NotifyResult{Detail: fmt.Sprintf("HTTP %d", status)}, nil
	}
	return entity.NotifyResult{Delivered: true, Detail: fmt.Sprintf("HTTP %d", status)}, nil
}
