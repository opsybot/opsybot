package discord

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
	defaultBase     = "https://discord.com/api/v10"
	maxResponseRead = 1 << 20
)

type Client struct {
	http      *http.Client
	base      string
	userAgent string
}

func New(cfg config.Discord) Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	base := cfg.BaseURL
	if base == "" {
		base = defaultBase
	}
	agent := cfg.UserAgent
	if agent == "" {
		agent = "opsybot"
	}
	return Client{http: &http.Client{Timeout: timeout}, base: strings.TrimRight(base, "/"), userAgent: agent}
}

func (c Client) do(ctx context.Context, token, method, path string, body any, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("encode discord request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return 0, fmt.Errorf("build discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bot "+token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("call discord %s: %w", path, err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return res.StatusCode, fmt.Errorf("discord %s responded %d: %s", path, res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return res.StatusCode, fmt.Errorf("decode discord %s: %w", path, err)
		}
	}
	return res.StatusCode, nil
}

type Application struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c Client) Me(ctx context.Context, token string) (Application, error) {
	var app Application
	_, err := c.do(ctx, token, http.MethodGet, "/applications/@me", nil, &app)
	return app, err
}

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c Client) Guild(ctx context.Context, token, guildID string) (Guild, error) {
	var g Guild
	_, err := c.do(ctx, token, http.MethodGet, "/guilds/"+guildID, nil, &g)
	return g, err
}

type Member struct {
	User struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
	} `json:"user"`
}

func (c Client) SearchMembers(ctx context.Context, token, guildID, query string) ([]Member, error) {
	var members []Member
	_, err := c.do(ctx, token, http.MethodGet, "/guilds/"+guildID+"/members/search?limit=10&query="+query, nil, &members)
	return members, err
}

type channel struct {
	ID string `json:"id"`
}

func (c Client) CreateDM(ctx context.Context, token, userID string) (string, error) {
	var ch channel
	_, err := c.do(ctx, token, http.MethodPost, "/users/@me/channels", map[string]any{"recipient_id": userID}, &ch)
	return ch.ID, err
}

type message struct {
	ID string `json:"id"`
}

func (c Client) CreateMessage(ctx context.Context, token, channelID, content string, components any) (string, error) {
	payload := map[string]any{"content": content}
	if components != nil {
		payload["components"] = components
	}
	var msg message
	_, err := c.do(ctx, token, http.MethodPost, "/channels/"+channelID+"/messages", payload, &msg)
	return msg.ID, err
}
