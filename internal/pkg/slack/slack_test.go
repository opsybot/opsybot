package slack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

func fakeIDToken(claims map[string]any) string {
	payload, _ := json.Marshal(claims)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func TestAuthorizeURL(t *testing.T) {
	c := New(config.Slack{ClientID: "cid-123"})
	raw := c.AuthorizeURL([]string{"chat:write", "im:write"}, "https://opsy.test/v1/chat/slack/oauth/callback", "state-abc")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	if q.Get("client_id") != "cid-123" {
		t.Errorf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("scope") != "chat:write,im:write" {
		t.Errorf("scope = %q, want comma-joined", q.Get("scope"))
	}
	if q.Get("state") != "state-abc" {
		t.Errorf("state = %q", q.Get("state"))
	}
	if q.Get("redirect_uri") != "https://opsy.test/v1/chat/slack/oauth/callback" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
}

func TestOAuthV2AccessSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth.v2.access" {
			t.Errorf("path = %q", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "cid" || pass != "csecret" {
			t.Errorf("basic auth = %q/%q ok=%v, want client credentials", user, pass, ok)
		}
		_ = r.ParseForm()
		if r.PostFormValue("code") != "the-code" {
			t.Errorf("code = %q", r.PostFormValue("code"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-abc","bot_user_id":"U0","scope":"chat:write,im:write","team":{"id":"T1","name":"Acme"}}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "csecret"})
	res, err := c.OAuthV2Access(context.Background(), "the-code", "https://opsy.test/cb")
	if err != nil {
		t.Fatalf("OAuthV2Access: %v", err)
	}
	if res.AccessToken != "xoxb-abc" || res.BotUserID != "U0" || res.TeamID != "T1" || res.TeamName != "Acme" {
		t.Errorf("result = %+v", res)
	}
}

func TestOAuthV2AccessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_code"}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "csecret"})
	_, err := c.OAuthV2Access(context.Background(), "bad", "https://opsy.test/cb")
	if err == nil || !strings.Contains(err.Error(), "invalid_code") {
		t.Fatalf("err = %v, want invalid_code", err)
	}
}

func TestOIDCAuthorizeURL(t *testing.T) {
	c := New(config.Slack{ClientID: "cid-9"})
	raw := c.OIDCAuthorizeURL([]string{"openid", "profile", "email"}, "https://opsy.test/v1/chat/slack/identity/callback", "st-1", "T42")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	if u.Host != "slack.com" || u.Path != "/openid/connect/authorize" {
		t.Errorf("endpoint = %s%s, want slack.com/openid/connect/authorize", u.Host, u.Path)
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q", q.Get("response_type"))
	}
	if q.Get("scope") != "openid profile email" {
		t.Errorf("scope = %q, want space-joined", q.Get("scope"))
	}
	if q.Get("client_id") != "cid-9" || q.Get("state") != "st-1" || q.Get("team") != "T42" {
		t.Errorf("client_id/state/team wrong: %q/%q/%q", q.Get("client_id"), q.Get("state"), q.Get("team"))
	}
}

func TestOpenIDConnectTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openid.connect.token" {
			t.Errorf("path = %q", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "cid" || pass != "csec" {
			t.Errorf("basic auth = %q/%q ok=%v", user, pass, ok)
		}
		token := fakeIDToken(map[string]any{
			"iss":                       "https://slack.com",
			"aud":                       "cid",
			"exp":                       time.Now().Add(time.Hour).Unix(),
			"https://slack.com/user_id": "U777",
			"https://slack.com/team_id": "T42",
			"name":                      "Vlad G",
			"email":                     "vlad@example.com",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"id_token":"` + token + `"}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "csec"})
	id, err := c.OpenIDConnectToken(context.Background(), "the-code", "https://opsy.test/cb")
	if err != nil {
		t.Fatalf("OpenIDConnectToken: %v", err)
	}
	if id.UserID != "U777" || id.TeamID != "T42" || id.Name != "Vlad G" || id.Email != "vlad@example.com" {
		t.Errorf("identity = %+v", id)
	}
}

func TestOpenIDConnectTokenRejectsWrongAudience(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := fakeIDToken(map[string]any{
			"iss": "https://slack.com", "aud": "someone-else",
			"exp": time.Now().Add(time.Hour).Unix(), "https://slack.com/user_id": "U1",
		})
		_, _ = w.Write([]byte(`{"ok":true,"id_token":"` + token + `"}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "csec"})
	_, err := c.OpenIDConnectToken(context.Background(), "code", "https://opsy.test/cb")
	if err == nil || !strings.Contains(err.Error(), "audience") {
		t.Fatalf("err = %v, want audience mismatch (id_token minted for another app must be rejected)", err)
	}
}

func TestPostMessageDecodesStringChannel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// chat.postMessage returns channel as a STRING, unlike conversations.open (object)
		_, _ = w.Write([]byte(`{"ok":true,"channel":"D0ABC","ts":"1700000000.000100"}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL})
	ts, err := c.PostMessage(context.Background(), "xoxb", "D0ABC", "hi", nil)
	if err != nil {
		t.Fatalf("PostMessage: %v (a string channel field must not fail decoding)", err)
	}
	if ts != "1700000000.000100" {
		t.Errorf("ts = %q", ts)
	}
}

func TestOpenIMDecodesObjectChannel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// conversations.open returns channel as an OBJECT
		_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"D0XYZ"}}`))
	}))
	defer srv.Close()

	c := New(config.Slack{BaseURL: srv.URL})
	ch, err := c.OpenIM(context.Background(), "xoxb", "U1")
	if err != nil {
		t.Fatalf("OpenIM: %v", err)
	}
	if ch != "D0XYZ" {
		t.Errorf("channel = %q, want D0XYZ", ch)
	}
}
