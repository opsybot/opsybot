package discord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
)

func TestAuthorizeURLBotInvite(t *testing.T) {
	c := New(config.Discord{ClientID: "app-1"})
	raw := c.AuthorizeURL([]string{"bot"}, "68624", "https://opsy.test/v1/chat/discord/oauth/callback", "st-1")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Host != "discord.com" || u.Path != "/oauth2/authorize" {
		t.Errorf("endpoint = %s%s", u.Host, u.Path)
	}
	q := u.Query()
	if q.Get("response_type") != "code" || q.Get("scope") != "bot" || q.Get("permissions") != "68624" {
		t.Errorf("bot invite params wrong: %v", q)
	}
	if q.Get("client_id") != "app-1" || q.Get("state") != "st-1" {
		t.Errorf("client_id/state wrong: %q/%q", q.Get("client_id"), q.Get("state"))
	}
}

func TestAuthorizeURLIdentifyOmitsPermissions(t *testing.T) {
	c := New(config.Discord{ClientID: "app-1"})
	raw := c.AuthorizeURL([]string{"identify"}, "", "https://opsy.test/cb", "st")
	q, _ := url.ParseQuery(strings.SplitN(raw, "?", 2)[1])
	if q.Get("scope") != "identify" {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	if q.Has("permissions") {
		t.Errorf("identify must not request bot permissions")
	}
}

func TestOAuthTokenAndCurrentUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "app-1" || pass != "sekret" {
				t.Errorf("basic auth = %q/%q ok=%v", user, pass, ok)
			}
			_ = r.ParseForm()
			if r.PostFormValue("code") != "the-code" {
				t.Errorf("code = %q", r.PostFormValue("code"))
			}
			_, _ = w.Write([]byte(`{"access_token":"bearer-xyz","token_type":"Bearer"}`))
		case "/users/@me":
			if r.Header.Get("Authorization") != "Bearer bearer-xyz" {
				t.Errorf("authorization = %q", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"id":"111","username":"vlad","global_name":"Vlad G"}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(config.Discord{BaseURL: srv.URL, ClientID: "app-1", ClientSecret: "sekret"})
	tok, err := c.OAuthToken(context.Background(), "the-code", "https://opsy.test/cb")
	if err != nil {
		t.Fatalf("OAuthToken: %v", err)
	}
	if tok != "bearer-xyz" {
		t.Fatalf("token = %q", tok)
	}
	u, err := c.CurrentUser(context.Background(), tok)
	if err != nil {
		t.Fatalf("CurrentUser: %v", err)
	}
	if u.ID != "111" || u.GlobalName != "Vlad G" {
		t.Errorf("user = %+v", u)
	}
}

func TestGuildChannels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/G9/channels" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bot the-bot" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`[{"id":"100","name":"general","type":0},{"id":"200","name":"incidents","type":0},{"id":"300","name":"Voice","type":2}]`))
	}))
	defer srv.Close()

	c := New(config.Discord{BaseURL: srv.URL})
	chans, err := c.GuildChannels(context.Background(), "the-bot", "G9")
	if err != nil {
		t.Fatalf("GuildChannels: %v", err)
	}
	if len(chans) != 3 || chans[1].Name != "incidents" || chans[1].ID != "200" || chans[1].Type != 0 {
		t.Errorf("channels = %+v", chans)
	}
}
