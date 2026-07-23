package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultLoginBase = "https://login.microsoftonline.com"
	defaultGraphBase = "https://graph.microsoft.com"
	defaultBotBase   = "https://smba.trafficmanager.net/teams"
	graphScope       = "https://graph.microsoft.com/.default"
	botScope         = "https://api.botframework.com/.default"
	maxResponseRead  = 1 << 20
)

type Client struct {
	http         *http.Client
	loginBase    string
	graphBase    string
	botBase      string
	userAgent    string
	appID        string
	appSecret    string
	tenantID     string
	catalogAppID string
}

func New(cfg config.Teams) Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return Client{
		http:         &http.Client{Timeout: timeout},
		loginBase:    base(cfg.LoginBaseURL, defaultLoginBase),
		graphBase:    base(cfg.GraphBaseURL, defaultGraphBase),
		botBase:      base(cfg.BotBaseURL, defaultBotBase),
		userAgent:    agent(cfg.UserAgent),
		appID:        cfg.AppID,
		appSecret:    cfg.AppSecret,
		tenantID:     cfg.TenantID,
		catalogAppID: cfg.CatalogAppID,
	}
}

func base(value, fallback string) string {
	if value == "" {
		value = fallback
	}
	return strings.TrimRight(value, "/")
}

func agent(value string) string {
	if value == "" {
		return "opsybot"
	}
	return value
}

func (c Client) Configured() bool {
	return c.appID != "" && c.appSecret != "" && c.tenantID != ""
}

func (c Client) token(ctx context.Context, scope string) (string, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.appID},
		"client_secret": {c.appSecret},
		"scope":         {scope},
	}
	endpoint := c.loginBase + "/" + c.tenantID + "/oauth2/v2.0/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build teams token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)
	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call teams token: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	var out struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode teams token (%d): %w", res.StatusCode, err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("teams token: %s", detail(out.Error, out.ErrorDescription))
	}
	return out.AccessToken, nil
}

type User struct {
	ID          string
	DisplayName string
}

func (c Client) LookupByEmail(ctx context.Context, email string) (User, error) {
	token, err := c.token(ctx, graphScope)
	if err != nil {
		return User{}, err
	}
	endpoint := c.graphBase + "/v1.0/users/" + url.PathEscape(email) + "?$select=id,displayName"
	var out struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	}
	if err := c.graph(ctx, http.MethodGet, endpoint, token, nil, &out); err != nil {
		return User{}, err
	}
	if out.ID == "" {
		return User{}, fmt.Errorf("teams user %q not found", email)
	}
	return User{ID: out.ID, DisplayName: out.DisplayName}, nil
}

func (c Client) InstallApp(ctx context.Context, userID string) error {
	if c.catalogAppID == "" {
		return nil
	}
	token, err := c.token(ctx, graphScope)
	if err != nil {
		return err
	}
	endpoint := c.graphBase + "/v1.0/users/" + url.PathEscape(userID) + "/teamwork/installedApps"
	body := map[string]any{
		"teamsApp@odata.bind": c.graphBase + "/v1.0/appCatalogs/teamsApps/" + c.catalogAppID,
	}
	err = c.graph(ctx, http.MethodPost, endpoint, token, body, nil)
	if err != nil && strings.Contains(err.Error(), "409") {
		return nil
	}
	return err
}

type Message struct {
	ConversationID string
	MessageID      string
}

func (c Client) SendDirect(ctx context.Context, userID, conversationID string, activity map[string]any) (Message, error) {
	if err := c.InstallApp(ctx, userID); err != nil {
		return Message{}, err
	}
	token, err := c.token(ctx, botScope)
	if err != nil {
		return Message{}, err
	}
	if conversationID == "" {
		conversationID, err = c.createConversation(ctx, token, userID)
		if err != nil {
			return Message{}, err
		}
	}
	endpoint := c.botBase + "/v3/conversations/" + url.PathEscape(conversationID) + "/activities"
	var out struct {
		ID string `json:"id"`
	}
	if err := c.bot(ctx, endpoint, token, activity, &out); err != nil {
		return Message{}, err
	}
	return Message{ConversationID: conversationID, MessageID: out.ID}, nil
}

func (c Client) createConversation(ctx context.Context, token, userID string) (string, error) {
	body := map[string]any{
		"bot":         map[string]any{"id": "28:" + c.appID},
		"members":     []map[string]any{{"id": userID}},
		"channelData": map[string]any{"tenant": map[string]any{"id": c.tenantID}},
		"tenantId":    c.tenantID,
		"isGroup":     false,
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := c.bot(ctx, c.botBase+"/v3/conversations", token, body, &out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("teams createConversation returned no id")
	}
	return out.ID, nil
}

func (c Client) Validate(ctx context.Context) error {
	_, err := c.token(ctx, graphScope)
	return err
}

func (c Client) graph(ctx context.Context, method, endpoint, token string, body, out any) error {
	return c.do(ctx, method, endpoint, token, body, out, "teams graph")
}

func (c Client) bot(ctx context.Context, endpoint, token string, body, out any) error {
	return c.do(ctx, http.MethodPost, endpoint, token, body, out, "teams bot")
}

func (c Client) do(ctx context.Context, method, endpoint, token string, body, out any, label string) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode %s request: %w", label, err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("build %s request: %w", label, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", label, err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s: %d %s", label, res.StatusCode, graphError(raw))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode %s (%d): %w", label, res.StatusCode, err)
		}
	}
	return nil
}

func graphError(raw []byte) string {
	var out struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &out); err == nil && out.Error.Message != "" {
		return detail(out.Error.Code, out.Error.Message)
	}
	return strings.TrimSpace(string(raw))
}

func detail(code, message string) string {
	switch {
	case code != "" && message != "":
		return code + ": " + message
	case message != "":
		return message
	default:
		return code
	}
}
