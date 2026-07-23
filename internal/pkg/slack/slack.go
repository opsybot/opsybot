package slack

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	defaultTimeout        = 10 * time.Second
	defaultBase           = "https://slack.com/api"
	authorizeEndpoint     = "https://slack.com/oauth/v2/authorize"
	oidcAuthorizeEndpoint = "https://slack.com/openid/connect/authorize"
	oidcIssuer            = "https://slack.com"
	maxResponseRead       = 1 << 20
)

type Client struct {
	http         *http.Client
	base         string
	userAgent    string
	clientID     string
	clientSecret string
}

func New(cfg config.Slack) Client {
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
	return Client{
		http:         &http.Client{Timeout: timeout},
		base:         strings.TrimRight(base, "/"),
		userAgent:    agent,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
	}
}

func (c Client) OAuthConfigured() bool {
	return c.clientID != "" && c.clientSecret != ""
}

func (c Client) AuthorizeURL(scopes []string, redirectURI, state string) string {
	q := url.Values{
		"client_id":    {c.clientID},
		"scope":        {strings.Join(scopes, ",")},
		"redirect_uri": {redirectURI},
		"state":        {state},
	}
	return authorizeEndpoint + "?" + q.Encode()
}

type channelRef struct {
	ID string
}

func (c *channelRef) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		c.ID = s
		return nil
	}
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	c.ID = obj.ID
	return nil
}

type apiResponse struct {
	OK      bool       `json:"ok"`
	Error   string     `json:"error"`
	Channel channelRef `json:"channel"`
	User    struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		RealName string `json:"real_name"`
		Deleted  bool   `json:"deleted"`
	} `json:"user"`
	Team   string `json:"team"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
	TS     string `json:"ts"`
}

func (c Client) call(ctx context.Context, token, method string, form url.Values, body any) (apiResponse, error) {
	var reader io.Reader
	contentType := "application/x-www-form-urlencoded"
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return apiResponse{}, fmt.Errorf("encode slack request: %w", err)
		}
		reader = bytes.NewReader(raw)
		contentType = "application/json; charset=utf-8"
	} else {
		reader = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/"+method, reader)
	if err != nil {
		return apiResponse{}, fmt.Errorf("build slack request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.http.Do(req)
	if err != nil {
		return apiResponse{}, fmt.Errorf("call slack %s: %w", method, err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	var out apiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return apiResponse{}, fmt.Errorf("decode slack %s (%d): %w", method, res.StatusCode, err)
	}
	return out, nil
}

type AuthTest struct {
	TeamID string
	Team   string
	UserID string
}

func (c Client) AuthTest(ctx context.Context, token string) (AuthTest, error) {
	res, err := c.call(ctx, token, "auth.test", url.Values{}, nil)
	if err != nil {
		return AuthTest{}, err
	}
	if !res.OK {
		return AuthTest{}, fmt.Errorf("slack auth.test: %s", res.Error)
	}
	return AuthTest{TeamID: res.TeamID, Team: res.Team, UserID: res.UserID}, nil
}

type User struct {
	ID       string
	Name     string
	RealName string
}

func (c Client) LookupByEmail(ctx context.Context, token, email string) (User, error) {
	res, err := c.call(ctx, token, "users.lookupByEmail", url.Values{"email": {email}}, nil)
	if err != nil {
		return User{}, err
	}
	if !res.OK {
		return User{}, fmt.Errorf("slack users.lookupByEmail: %s", res.Error)
	}
	return User{ID: res.User.ID, Name: res.User.Name, RealName: res.User.RealName}, nil
}

func (c Client) OpenIM(ctx context.Context, token, userID string) (string, error) {
	res, err := c.call(ctx, token, "conversations.open", nil, map[string]any{"users": userID})
	if err != nil {
		return "", err
	}
	if !res.OK {
		return "", fmt.Errorf("slack conversations.open: %s", res.Error)
	}
	return res.Channel.ID, nil
}

type OAuthResult struct {
	AccessToken string
	BotUserID   string
	TeamID      string
	TeamName    string
	Scope       string
}

type oauthResponse struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

func (c Client) OAuthV2Access(ctx context.Context, code, redirectURI string) (OAuthResult, error) {
	form := url.Values{"code": {code}, "redirect_uri": {redirectURI}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/oauth.v2.access", strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthResult{}, fmt.Errorf("build slack oauth.v2.access: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)
	req.SetBasicAuth(c.clientID, c.clientSecret)
	res, err := c.http.Do(req)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("call slack oauth.v2.access: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	var out oauthResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return OAuthResult{}, fmt.Errorf("decode slack oauth.v2.access (%d): %w", res.StatusCode, err)
	}
	if !out.OK {
		return OAuthResult{}, fmt.Errorf("slack oauth.v2.access: %s", out.Error)
	}
	return OAuthResult{
		AccessToken: out.AccessToken, BotUserID: out.BotUserID,
		TeamID: out.Team.ID, TeamName: out.Team.Name, Scope: out.Scope,
	}, nil
}

func (c Client) OIDCAuthorizeURL(scopes []string, redirectURI, state, teamID string) string {
	q := url.Values{
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"client_id":     {c.clientID},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	if teamID != "" {
		q.Set("team", teamID)
	}
	return oidcAuthorizeEndpoint + "?" + q.Encode()
}

type OIDCIdentity struct {
	UserID string
	TeamID string
	Name   string
	Email  string
}

type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*a = audience{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

type oidcClaims struct {
	Iss    string   `json:"iss"`
	Aud    audience `json:"aud"`
	Exp    int64    `json:"exp"`
	UserID string   `json:"https://slack.com/user_id"`
	TeamID string   `json:"https://slack.com/team_id"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
}

func (c Client) OpenIDConnectToken(ctx context.Context, code, redirectURI string) (OIDCIdentity, error) {
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURI}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/openid.connect.token", strings.NewReader(form.Encode()))
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("build slack openid.connect.token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)
	req.SetBasicAuth(c.clientID, c.clientSecret)
	res, err := c.http.Do(req)
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("call slack openid.connect.token: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseRead))
	var out struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return OIDCIdentity{}, fmt.Errorf("decode slack openid.connect.token (%d): %w", res.StatusCode, err)
	}
	if !out.OK {
		return OIDCIdentity{}, fmt.Errorf("slack openid.connect.token: %s", out.Error)
	}
	claims, err := c.decodeIDToken(out.IDToken)
	if err != nil {
		return OIDCIdentity{}, err
	}
	return OIDCIdentity{UserID: claims.UserID, TeamID: claims.TeamID, Name: claims.Name, Email: claims.Email}, nil
}

func (c Client) decodeIDToken(idToken string) (oidcClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return oidcClaims{}, fmt.Errorf("slack id_token malformed")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return oidcClaims{}, fmt.Errorf("decode slack id_token: %w", err)
	}
	var claims oidcClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return oidcClaims{}, fmt.Errorf("parse slack id_token claims: %w", err)
	}
	if claims.Iss != oidcIssuer {
		return oidcClaims{}, fmt.Errorf("slack id_token issuer %q invalid", claims.Iss)
	}
	if !slices.Contains([]string(claims.Aud), c.clientID) {
		return oidcClaims{}, fmt.Errorf("slack id_token audience mismatch")
	}
	if claims.Exp > 0 && time.Now().Unix() >= claims.Exp {
		return oidcClaims{}, fmt.Errorf("slack id_token expired")
	}
	if claims.UserID == "" {
		return oidcClaims{}, fmt.Errorf("slack id_token missing user id")
	}
	return claims, nil
}

func (c Client) PostMessage(ctx context.Context, token, channel, text string, blocks any) (string, error) {
	payload := map[string]any{"channel": channel, "text": text}
	if blocks != nil {
		payload["blocks"] = blocks
	}
	res, err := c.call(ctx, token, "chat.postMessage", nil, payload)
	if err != nil {
		return "", err
	}
	if !res.OK {
		return "", fmt.Errorf("slack chat.postMessage: %s", res.Error)
	}
	return res.TS, nil
}
