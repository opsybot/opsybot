package teams

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
)

func newServer(t *testing.T, hits map[string]int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/oauth2/v2.0/token"):
			hits["token"]++
			_, _ = w.Write([]byte(`{"access_token":"tok-123","token_type":"Bearer","expires_in":3599}`))
		case strings.Contains(path, "/teamwork/installedApps"):
			hits["install"]++
			w.WriteHeader(http.StatusCreated)
		case strings.Contains(path, "/v1.0/users/"):
			hits["lookup"]++
			_, _ = w.Write([]byte(`{"id":"aad-1","displayName":"Vlad Gorokhov"}`))
		case strings.HasSuffix(path, "/v3/conversations"):
			hits["createConversation"]++
			_, _ = w.Write([]byte(`{"id":"conv-1"}`))
		case strings.Contains(path, "/v3/conversations/") && strings.HasSuffix(path, "/activities"):
			hits["activity"]++
			_, _ = w.Write([]byte(`{"id":"msg-1"}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, path)
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
}

func testClient(url string) Client {
	return New(config.Teams{
		AppID: "app-1", AppSecret: "secret", TenantID: "tenant-1", CatalogAppID: "catalog-1",
		LoginBaseURL: url, GraphBaseURL: url, BotBaseURL: url,
	})
}

func TestConfiguredRequiresAppCredsAndTenant(t *testing.T) {
	if New(config.Teams{AppID: "a", AppSecret: "s"}).Configured() {
		t.Fatal("a Teams client with no tenant must not be configured")
	}
	if !New(config.Teams{AppID: "a", AppSecret: "s", TenantID: "t"}).Configured() {
		t.Fatal("app id + secret + tenant should be configured")
	}
}

func TestValidateAcquiresToken(t *testing.T) {
	hits := map[string]int{}
	srv := newServer(t, hits)
	defer srv.Close()
	if err := testClient(srv.URL).Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if hits["token"] != 1 {
		t.Fatalf("Validate should acquire exactly one token, got %d", hits["token"])
	}
}

func TestValidateRejectsBadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"bad secret"}`))
	}))
	defer srv.Close()
	if err := testClient(srv.URL).Validate(context.Background()); err == nil {
		t.Fatal("Validate must fail when the token endpoint rejects the credentials")
	}
}

func TestLookupByEmailResolvesAADUser(t *testing.T) {
	hits := map[string]int{}
	srv := newServer(t, hits)
	defer srv.Close()
	user, err := testClient(srv.URL).LookupByEmail(context.Background(), "vlad@acme.test")
	if err != nil {
		t.Fatalf("LookupByEmail: %v", err)
	}
	if user.ID != "aad-1" || user.DisplayName != "Vlad Gorokhov" {
		t.Fatalf("resolved user = %+v", user)
	}
}

func TestSendDirectInstallsCreatesConversationAndPosts(t *testing.T) {
	hits := map[string]int{}
	srv := newServer(t, hits)
	defer srv.Close()
	msg, err := testClient(srv.URL).SendDirect(context.Background(), "aad-1", "", map[string]any{"type": "message"})
	if err != nil {
		t.Fatalf("SendDirect: %v", err)
	}
	if msg.ConversationID != "conv-1" || msg.MessageID != "msg-1" {
		t.Fatalf("message = %+v", msg)
	}
	if hits["install"] != 1 || hits["createConversation"] != 1 || hits["activity"] != 1 {
		t.Fatalf("expected install+create+activity once each, got %+v", hits)
	}
}

func TestSendDirectReusesExistingConversation(t *testing.T) {
	hits := map[string]int{}
	srv := newServer(t, hits)
	defer srv.Close()
	msg, err := testClient(srv.URL).SendDirect(context.Background(), "aad-1", "conv-existing", map[string]any{"type": "message"})
	if err != nil {
		t.Fatalf("SendDirect: %v", err)
	}
	if msg.ConversationID != "conv-existing" {
		t.Fatalf("a known conversation must be reused, got %q", msg.ConversationID)
	}
	if hits["createConversation"] != 0 {
		t.Fatalf("createConversation must be skipped when a conversation id is known, got %d", hits["createConversation"])
	}
	if hits["activity"] != 1 {
		t.Fatalf("expected one activity post, got %d", hits["activity"])
	}
}
