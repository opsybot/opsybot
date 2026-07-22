package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	defaultTimeout  = 10 * time.Second
	maxResponseRead = 4096
)

type Client struct {
	http      *http.Client
	userAgent string
}

func New(cfg config.Webhook) Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	agent := cfg.UserAgent
	if agent == "" {
		agent = "opsybot"
	}
	return Client{
		http:      &http.Client{Timeout: timeout},
		userAgent: agent,
	}
}

func (c Client) Post(ctx context.Context, url, signature string, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if signature != "" {
		req.Header.Set("X-Opsy-Signature", signature)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("deliver webhook: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, maxResponseRead))
	return res.StatusCode, nil
}
