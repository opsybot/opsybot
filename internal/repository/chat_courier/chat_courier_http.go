package chat_courier

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/discord"
	"github.com/opsybot/opsybot/internal/pkg/slack"
	"github.com/opsybot/opsybot/internal/pkg/teams"
	"github.com/opsybot/opsybot/internal/pkg/telegram"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	slack    slack.Client
	discord  discord.Client
	telegram telegram.Client
	teams    teams.Client
}

func New(slackClient slack.Client, discordClient discord.Client, telegramClient telegram.Client, teamsClient teams.Client) repository.ChatCourier {
	return &repo{slack: slackClient, discord: discordClient, telegram: telegramClient, teams: teamsClient}
}

func (r *repo) Send(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error) {
	switch in.Provider {
	case entity.ChatProviderSlack:
		return r.sendSlack(ctx, in)
	case entity.ChatProviderDiscord:
		return r.sendDiscord(ctx, in)
	case entity.ChatProviderTelegram:
		return r.sendTelegram(ctx, in)
	case entity.ChatProviderTeams:
		return r.sendTeams(ctx, in)
	default:
		return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "chat provider not supported yet"}}, nil
	}
}

func (r *repo) sendTeams(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error) {
	msg, err := r.teams.SendDirect(ctx, in.ProviderUserID, in.DMChannelID, teamsActivity(in))
	if err != nil {
		return entity.ChatSendResult{DMChannelID: in.DMChannelID, Result: entity.NotifyResult{Detail: err.Error()}}, nil
	}
	return entity.ChatSendResult{DMChannelID: msg.ConversationID, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Microsoft Teams", MessageID: msg.MessageID}}, nil
}

func teamsActivity(in entity.ChatDelivery) map[string]any {
	body := []map[string]any{
		{"type": "TextBlock", "text": in.Page.Subject(), "weight": "Bolder", "wrap": true},
		{"type": "TextBlock", "text": strings.Join(in.Page.BodyLines(), "  ·  "), "wrap": true, "isSubtle": true},
	}
	actions := []map[string]any{}
	base := strings.TrimRight(in.BaseURL, "/")
	if in.AckToken != "" && base != "" {
		actions = append(actions, map[string]any{"type": "Action.OpenUrl", "title": "Acknowledge", "url": base + "/v1/act/" + in.AckToken})
	}
	if in.ResolveToken != "" && base != "" {
		actions = append(actions, map[string]any{"type": "Action.OpenUrl", "title": "Resolve", "url": base + "/v1/act/" + in.ResolveToken})
	}
	if in.Page.AlertURL != "" {
		actions = append(actions, map[string]any{"type": "Action.OpenUrl", "title": "Open alert", "url": in.Page.AlertURL})
	}
	card := map[string]any{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.4",
		"body":    body,
	}
	if len(actions) > 0 {
		card["actions"] = actions
	}
	return map[string]any{
		"type": "message",
		"attachments": []map[string]any{
			{"contentType": "application/vnd.microsoft.card.adaptive", "content": card},
		},
	}
}

func (r *repo) sendTelegram(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error) {
	chatID := in.DMChannelID
	if chatID == "" {
		chatID = in.ProviderUserID
	}
	mid, err := r.telegram.SendMessage(ctx, in.BotToken, chatID, in.Page.PlainText(), telegramKeyboard(in))
	if err != nil {
		return entity.ChatSendResult{DMChannelID: chatID, Result: entity.NotifyResult{Detail: err.Error()}}, nil
	}
	return entity.ChatSendResult{DMChannelID: chatID, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Telegram", MessageID: strconv.FormatInt(mid, 10)}}, nil
}

func telegramKeyboard(in entity.ChatDelivery) any {
	row := []map[string]any{}
	if in.AckToken != "" {
		row = append(row, map[string]any{"text": "Acknowledge", "callback_data": in.AckToken})
	}
	if in.ResolveToken != "" {
		row = append(row, map[string]any{"text": "Resolve", "callback_data": in.ResolveToken})
	}
	if in.Page.AlertURL != "" {
		row = append(row, map[string]any{"text": "Open alert", "url": in.Page.AlertURL})
	}
	if len(row) == 0 {
		return nil
	}
	return map[string]any{"inline_keyboard": [][]map[string]any{row}}
}

func (r *repo) sendSlack(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error) {
	dm := in.DMChannelID
	if dm == "" {
		opened, err := r.slack.OpenIM(ctx, in.BotToken, in.ProviderUserID)
		if err != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		dm = opened
	}
	ts, err := r.slack.PostMessage(ctx, in.BotToken, dm, in.Page.Subject(), slackBlocks(in))
	if err != nil {
		return entity.ChatSendResult{DMChannelID: dm, Result: entity.NotifyResult{Detail: err.Error()}}, nil
	}
	return entity.ChatSendResult{DMChannelID: dm, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Slack", MessageID: ts}}, nil
}

func slackBlocks(in entity.ChatDelivery) []map[string]any {
	blocks := []map[string]any{
		{"type": "section", "text": map[string]any{"type": "mrkdwn", "text": "*" + in.Page.Subject() + "*"}},
		{"type": "context", "elements": []map[string]any{
			{"type": "mrkdwn", "text": strings.Join(in.Page.BodyLines(), "  ·  ")},
		}},
	}
	elements := []map[string]any{}
	if in.AckToken != "" {
		elements = append(elements, map[string]any{
			"type": "button", "text": map[string]any{"type": "plain_text", "text": "Acknowledge"},
			"style": "primary", "action_id": "ack", "value": in.AckToken,
		})
	}
	if in.ResolveToken != "" {
		elements = append(elements, map[string]any{
			"type": "button", "text": map[string]any{"type": "plain_text", "text": "Resolve"},
			"action_id": "resolve", "value": in.ResolveToken,
		})
	}
	if in.Page.AlertURL != "" {
		elements = append(elements, map[string]any{
			"type": "button", "text": map[string]any{"type": "plain_text", "text": "Open alert"}, "url": in.Page.AlertURL,
		})
	}
	if len(elements) > 0 {
		blocks = append(blocks, map[string]any{"type": "actions", "elements": elements})
	}
	return blocks
}

func (r *repo) sendDiscord(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error) {
	dm := in.DMChannelID
	if dm == "" {
		created, err := r.discord.CreateDM(ctx, in.BotToken, in.ProviderUserID)
		if err != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		dm = created
	}
	id, err := r.discord.CreateMessage(ctx, in.BotToken, dm, in.Page.PlainText(), discordComponents(in))
	if err != nil {
		return entity.ChatSendResult{DMChannelID: dm, Result: entity.NotifyResult{Detail: err.Error()}}, nil
	}
	return entity.ChatSendResult{DMChannelID: dm, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Discord", MessageID: id}}, nil
}

func (r *repo) Validate(ctx context.Context, provider entity.ChatProvider, token, externalID string) (entity.ChatValidation, error) {
	switch provider {
	case entity.ChatProviderSlack:
		auth, err := r.slack.AuthTest(ctx, token)
		if err != nil {
			return entity.ChatValidation{}, entity.ErrChatConnectionInvalid
		}
		return entity.ChatValidation{ExternalID: auth.TeamID, ExternalName: auth.Team, BotUserID: auth.UserID}, nil
	case entity.ChatProviderDiscord:
		app, err := r.discord.Me(ctx, token)
		if err != nil {
			return entity.ChatValidation{}, entity.ErrChatConnectionInvalid
		}
		name := app.Name
		if externalID != "" {
			if guild, err := r.discord.Guild(ctx, token, externalID); err == nil {
				name = guild.Name
			}
		}
		return entity.ChatValidation{ExternalID: externalID, ExternalName: name, BotUserID: app.ID}, nil
	case entity.ChatProviderTelegram:
		bot, err := r.telegram.Me(ctx, token)
		if err != nil {
			return entity.ChatValidation{}, entity.ErrChatConnectionInvalid
		}
		id := strconv.FormatInt(bot.ID, 10)
		return entity.ChatValidation{ExternalID: id, ExternalName: bot.Username, BotUserID: id}, nil
	case entity.ChatProviderTeams:
		if !r.teams.Configured() {
			return entity.ChatValidation{}, entity.ErrChatProviderNotConfigured
		}
		if err := r.teams.Validate(ctx); err != nil {
			return entity.ChatValidation{}, entity.ErrChatConnectionInvalid
		}
		return entity.ChatValidation{ExternalID: externalID, ExternalName: "Microsoft Teams"}, nil
	default:
		return entity.ChatValidation{}, entity.ErrChatConnectionInvalid
	}
}

func (r *repo) SetWebhook(ctx context.Context, provider entity.ChatProvider, token, webhookURL, secret string) error {
	switch provider {
	case entity.ChatProviderTelegram:
		return r.telegram.SetWebhook(ctx, token, webhookURL, secret)
	default:
		return entity.ErrChatOAuthUnsupported
	}
}

func (r *repo) AnswerCallback(ctx context.Context, provider entity.ChatProvider, token, callbackID, text string) error {
	switch provider {
	case entity.ChatProviderTelegram:
		return r.telegram.AnswerCallbackQuery(ctx, token, callbackID, text)
	default:
		return nil
	}
}

func (r *repo) LookupUser(ctx context.Context, provider entity.ChatProvider, token, externalID, email string) (entity.ChatUser, error) {
	switch provider {
	case entity.ChatProviderSlack:
		user, err := r.slack.LookupByEmail(ctx, token, email)
		if err != nil {
			return entity.ChatUser{}, entity.ErrChatConnectionInvalid
		}
		return entity.ChatUser{ProviderUserID: user.ID, Handle: user.Name}, nil
	case entity.ChatProviderDiscord:
		query := email
		if at := strings.IndexByte(email, '@'); at > 0 {
			query = email[:at]
		}
		members, err := r.discord.SearchMembers(ctx, token, externalID, query)
		if err != nil || len(members) == 0 {
			return entity.ChatUser{}, entity.ErrChatConnectionInvalid
		}
		u := members[0].User
		handle := u.GlobalName
		if handle == "" {
			handle = u.Username
		}
		return entity.ChatUser{ProviderUserID: u.ID, Handle: handle}, nil
	case entity.ChatProviderTeams:
		user, err := r.teams.LookupByEmail(ctx, email)
		if err != nil {
			return entity.ChatUser{}, entity.ErrChatConnectionInvalid
		}
		return entity.ChatUser{ProviderUserID: user.ID, Handle: user.DisplayName}, nil
	default:
		return entity.ChatUser{}, entity.ErrChatConnectionInvalid
	}
}

func (r *repo) SendToChannel(ctx context.Context, provider entity.ChatProvider, token, guildID, channel, text string) (entity.ChatSendResult, error) {
	switch provider {
	case entity.ChatProviderSlack:
		target := strings.TrimPrefix(channel, "#")
		ts, err := r.slack.PostMessage(ctx, token, target, text, nil)
		if err != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		return entity.ChatSendResult{DMChannelID: target, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Slack", MessageID: ts}}, nil
	case entity.ChatProviderDiscord:
		if guildID == "" {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "Connect Discord to a server first."}}, nil
		}
		channels, err := r.discord.GuildChannels(ctx, token, guildID)
		if err != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		want := strings.ToLower(strings.TrimPrefix(channel, "#"))
		channelID := ""
		for _, gc := range channels {
			if (gc.Type == 0 || gc.Type == 5) && strings.ToLower(gc.Name) == want {
				channelID = gc.ID
				break
			}
		}
		if channelID == "" {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "Channel " + channel + " was not found in the server."}}, nil
		}
		id, err := r.discord.CreateMessage(ctx, token, channelID, text, nil)
		if err != nil {
			return entity.ChatSendResult{DMChannelID: channelID, Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		return entity.ChatSendResult{DMChannelID: channelID, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Discord", MessageID: id}}, nil
	case entity.ChatProviderTelegram:
		if channel == "" {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "Set a Telegram channel (@name or id) as the announcement channel first."}}, nil
		}
		mid, err := r.telegram.SendMessage(ctx, token, channel, text, nil)
		if err != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: err.Error()}}, nil
		}
		return entity.ChatSendResult{DMChannelID: channel, Result: entity.NotifyResult{Delivered: true, Detail: "delivered to Telegram", MessageID: strconv.FormatInt(mid, 10)}}, nil
	default:
		return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "chat provider not supported yet"}}, nil
	}
}

func (r *repo) AuthorizeURL(ctx context.Context, provider entity.ChatProvider, scopes []string, redirectURI, state string) (string, error) {
	switch provider {
	case entity.ChatProviderSlack:
		if !r.slack.OAuthConfigured() {
			return "", entity.ErrChatProviderNotConfigured
		}
		return r.slack.AuthorizeURL(scopes, redirectURI, state), nil
	case entity.ChatProviderDiscord:
		if !r.discord.OAuthConfigured() {
			return "", entity.ErrChatProviderNotConfigured
		}
		return r.discord.AuthorizeURL(scopes, entity.DiscordBotPermissions, redirectURI, state), nil
	default:
		return "", entity.ErrChatOAuthUnsupported
	}
}

func (r *repo) IdentityAuthorizeURL(ctx context.Context, provider entity.ChatProvider, scopes []string, redirectURI, state, teamID string) (string, error) {
	switch provider {
	case entity.ChatProviderSlack:
		if !r.slack.OAuthConfigured() {
			return "", entity.ErrChatProviderNotConfigured
		}
		return r.slack.OIDCAuthorizeURL(scopes, redirectURI, state, teamID), nil
	case entity.ChatProviderDiscord:
		if !r.discord.OAuthConfigured() {
			return "", entity.ErrChatProviderNotConfigured
		}
		return r.discord.AuthorizeURL(scopes, "", redirectURI, state), nil
	default:
		return "", entity.ErrChatOAuthUnsupported
	}
}

func (r *repo) ExchangeIdentity(ctx context.Context, provider entity.ChatProvider, code, redirectURI string) (entity.ChatIdentityResult, error) {
	switch provider {
	case entity.ChatProviderSlack:
		res, err := r.slack.OpenIDConnectToken(ctx, code, redirectURI)
		if err != nil {
			return entity.ChatIdentityResult{}, fmt.Errorf("%w: %v", entity.ErrChatOAuthExchange, err)
		}
		return entity.ChatIdentityResult{
			ProviderUserID: res.UserID, TeamID: res.TeamID, Handle: res.Name, Email: res.Email,
		}, nil
	case entity.ChatProviderDiscord:
		accessToken, err := r.discord.OAuthToken(ctx, code, redirectURI)
		if err != nil {
			return entity.ChatIdentityResult{}, fmt.Errorf("%w: %v", entity.ErrChatOAuthExchange, err)
		}
		user, err := r.discord.CurrentUser(ctx, accessToken)
		if err != nil {
			return entity.ChatIdentityResult{}, fmt.Errorf("%w: %v", entity.ErrChatOAuthExchange, err)
		}
		handle := user.GlobalName
		if handle == "" {
			handle = user.Username
		}
		return entity.ChatIdentityResult{ProviderUserID: user.ID, Handle: handle}, nil
	default:
		return entity.ChatIdentityResult{}, entity.ErrChatOAuthUnsupported
	}
}

func (r *repo) ExchangeOAuth(ctx context.Context, provider entity.ChatProvider, code, redirectURI string) (entity.ChatOAuthResult, error) {
	switch provider {
	case entity.ChatProviderSlack:
		res, err := r.slack.OAuthV2Access(ctx, code, redirectURI)
		if err != nil {
			return entity.ChatOAuthResult{}, fmt.Errorf("%w: %v", entity.ErrChatOAuthExchange, err)
		}
		var scopes []string
		if res.Scope != "" {
			scopes = strings.Split(res.Scope, ",")
		}
		return entity.ChatOAuthResult{
			ExternalID: res.TeamID, ExternalName: res.TeamName, BotUserID: res.BotUserID,
			BotToken: res.AccessToken, Scopes: scopes,
		}, nil
	default:
		return entity.ChatOAuthResult{}, entity.ErrChatOAuthUnsupported
	}
}

func discordComponents(in entity.ChatDelivery) []map[string]any {
	buttons := []map[string]any{}
	if in.AckToken != "" {
		buttons = append(buttons, map[string]any{"type": 2, "style": 1, "label": "Acknowledge", "custom_id": in.AckToken})
	}
	if in.ResolveToken != "" {
		buttons = append(buttons, map[string]any{"type": 2, "style": 2, "label": "Resolve", "custom_id": in.ResolveToken})
	}
	if in.Page.AlertURL != "" {
		buttons = append(buttons, map[string]any{"type": 2, "style": 5, "label": "Open alert", "url": in.Page.AlertURL})
	}
	if len(buttons) == 0 {
		return nil
	}
	return []map[string]any{{"type": 1, "components": buttons}}
}
