package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	defaultTimeout  = 10 * time.Second
	defaultBase     = "https://api.telegram.org"
	maxResponseRead = 1 << 20
)

type Client struct {
	http      *http.Client
	base      string
	userAgent string
}

func New(cfg config.Telegram) Client {
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

func (c Client) call(ctx context.Context, token, method string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode telegram %s: %w", method, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/bot"+token+"/"+method, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build telegram %s: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call telegram %s: %w", method, err)
	}
	defer func() { _ = res.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	var envelope struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode telegram %s (%d): %w", method, res.StatusCode, err)
	}
	if !envelope.OK {
		return fmt.Errorf("telegram %s: %s", method, envelope.Description)
	}
	if out != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, out); err != nil {
			return fmt.Errorf("decode telegram %s result: %w", method, err)
		}
	}
	return nil
}

type Bot struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

func (c Client) Me(ctx context.Context, token string) (Bot, error) {
	var bot Bot
	err := c.call(ctx, token, "getMe", map[string]any{}, &bot)
	return bot, err
}

func (c Client) SetWebhook(ctx context.Context, token, url, secret string) error {
	return c.call(ctx, token, "setWebhook", map[string]any{
		"url":                  url,
		"secret_token":         secret,
		"allowed_updates":      []string{"message", "callback_query"},
		"drop_pending_updates": true,
	}, nil)
}

func (c Client) SendMessage(ctx context.Context, token, chatID, text string, replyMarkup any) (int64, error) {
	payload := map[string]any{"chat_id": chatIDValue(chatID), "text": text}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	var msg struct {
		MessageID int64 `json:"message_id"`
	}
	if err := c.call(ctx, token, "sendMessage", payload, &msg); err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}

func (c Client) AnswerCallbackQuery(ctx context.Context, token, callbackID, text string) error {
	return c.call(ctx, token, "answerCallbackQuery", map[string]any{"callback_query_id": callbackID, "text": text}, nil)
}

func chatIDValue(chatID string) any {
	if strings.HasPrefix(chatID, "@") {
		return chatID
	}
	if n, err := strconv.ParseInt(chatID, 10, 64); err == nil {
		return n
	}
	return chatID
}
