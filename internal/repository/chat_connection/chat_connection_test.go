package chat_connection

import (
	"testing"

	"github.com/opsybot/opsybot/internal/entity"
)

func TestEnvTokenPerProvider(t *testing.T) {
	// The env-token fallback must cover Telegram — otherwise BotToken(telegram)
	// returns "" and every delivery/test path reports "not connected".
	r := &repo{slackToken: "xoxb", discToken: "disc", telegramToken: "tg"}
	cases := map[entity.ChatProvider]string{
		entity.ChatProviderSlack:    "xoxb",
		entity.ChatProviderDiscord:  "disc",
		entity.ChatProviderTelegram: "tg",
		entity.ChatProviderTeams:    "",
	}
	for provider, want := range cases {
		if got := r.envToken(provider); got != want {
			t.Errorf("envToken(%s) = %q, want %q", provider, got, want)
		}
	}
}
