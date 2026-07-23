package ntfy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	defaultTimeout  = 10 * time.Second
	defaultServer   = "https://ntfy.sh"
	maxResponseRead = 8192
)

type Client struct {
	http          *http.Client
	userAgent     string
	defaultServer string
}

func New(cfg config.Ntfy) Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	agent := cfg.UserAgent
	if agent == "" {
		agent = "opsybot"
	}
	server := cfg.DefaultServer
	if server == "" {
		server = defaultServer
	}
	return Client{http: &http.Client{Timeout: timeout}, userAgent: agent, defaultServer: strings.TrimRight(server, "/")}
}

type Action struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Method string `json:"method,omitempty"`
	Clear  bool   `json:"clear,omitempty"`
}

type Message struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title,omitempty"`
	Message  string   `json:"message,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Click    string   `json:"click,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
}

type Published struct {
	ID string `json:"id"`
}

func (c Client) Publish(ctx context.Context, server, token string, msg Message) (Published, error) {
	base := c.defaultServer
	if server != "" {
		base = strings.TrimRight(server, "/")
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return Published{}, fmt.Errorf("encode ntfy message: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, bytes.NewReader(body))
	if err != nil {
		return Published{}, fmt.Errorf("build ntfy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return Published{}, fmt.Errorf("publish ntfy: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return Published{}, fmt.Errorf("ntfy responded %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out Published
	_ = json.Unmarshal(raw, &out)
	return out, nil
}
