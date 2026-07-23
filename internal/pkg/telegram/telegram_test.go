package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
)

func TestMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottok/getMe" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"id":42,"username":"opsy_bot","first_name":"Opsy"}}`))
	}))
	defer srv.Close()

	c := New(config.Telegram{BaseURL: srv.URL})
	bot, err := c.Me(context.Background(), "tok")
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if bot.ID != 42 || bot.Username != "opsy_bot" {
		t.Errorf("bot = %+v", bot)
	}
}

func TestSendMessageNumericAndChannel(t *testing.T) {
	var gotChatIDs []any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		gotChatIDs = append(gotChatIDs, payload["chat_id"])
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":7}}`))
	}))
	defer srv.Close()

	c := New(config.Telegram{BaseURL: srv.URL})
	if _, err := c.SendMessage(context.Background(), "tok", "12345", "hi", nil); err != nil {
		t.Fatalf("SendMessage numeric: %v", err)
	}
	if _, err := c.SendMessage(context.Background(), "tok", "@ incidents", "hi", nil); err != nil {
		t.Fatalf("SendMessage channel: %v", err)
	}
	// numeric chat_id decodes to a JSON number; @channel stays a string
	if n, ok := gotChatIDs[0].(float64); !ok || n != 12345 {
		t.Errorf("numeric chat_id = %v (%T), want 12345 as number", gotChatIDs[0], gotChatIDs[0])
	}
	if s, ok := gotChatIDs[1].(string); !ok || !strings.HasPrefix(s, "@") {
		t.Errorf("channel chat_id = %v, want @-string", gotChatIDs[1])
	}
}

func TestSetWebhookError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"description":"bad webhook"}`))
	}))
	defer srv.Close()

	c := New(config.Telegram{BaseURL: srv.URL})
	err := c.SetWebhook(context.Background(), "tok", "https://opsy.test/hook", "sec")
	if err == nil || !strings.Contains(err.Error(), "bad webhook") {
		t.Fatalf("err = %v, want bad webhook", err)
	}
}
