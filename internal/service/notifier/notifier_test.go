package notifier

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/chat_connection"
	"github.com/opsybot/opsybot/internal/repository/chat_courier"
	"github.com/opsybot/opsybot/internal/repository/chat_identity"
	"github.com/opsybot/opsybot/internal/repository/mailer"
	"github.com/opsybot/opsybot/internal/repository/ntfy"
	"github.com/opsybot/opsybot/internal/repository/pager"
)

type harness struct {
	srv     *srv
	mailer  *mailer.MockMailer
	pager   *pager.MockPager
	ntfy    *ntfy.MockNtfy
	conns   *chat_connection.MockChatConnection
	ids     *chat_identity.MockChatIdentity
	courier *chat_courier.MockChatCourier
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		mailer:  mailer.NewMockMailer(ctrl),
		pager:   pager.NewMockPager(ctrl),
		ntfy:    ntfy.NewMockNtfy(ctrl),
		conns:   chat_connection.NewMockChatConnection(ctrl),
		ids:     chat_identity.NewMockChatIdentity(ctrl),
		courier: chat_courier.NewMockChatCourier(ctrl),
	}
	h.srv = &srv{
		mailer: h.mailer, pager: h.pager, ntfy: h.ntfy,
		chatConns: h.conns, chatIDs: h.ids, chatCourier: h.courier,
		cfg: config.Auth{BaseURL: "https://opsy.test"},
	}
	return h
}

func (h *harness) expectChatSend(health entity.ChatHealth, result entity.ChatSendResult) {
	h.conns.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderSlack).
		Return(entity.ChatConnection{ID: "c1", Health: health}, nil)
	h.conns.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderSlack).Return("xoxb", nil)
	h.ids.EXPECT().GetForUser(gomock.Any(), "c1", "u1").
		Return(entity.ChatIdentity{ProviderUserID: "U1", DMChannelID: "d1", Verified: true}, nil)
	h.courier.EXPECT().Send(gomock.Any(), gomock.Any()).Return(result, nil)
}

func slackTarget() entity.NotifyTarget {
	return entity.NotifyTarget{WorkspaceID: "ws-1", UserID: "u1", Channel: entity.ChannelTypeSlack}
}

func TestChatDeliveryFailureFlipsHealthToFailing(t *testing.T) {
	h := newHarness(t)
	h.expectChatSend(entity.ChatHealthy, entity.ChatSendResult{
		DMChannelID: "d1", Result: entity.NotifyResult{Delivered: false, Detail: "invalid_auth"},
	})
	h.conns.EXPECT().SetHealth(gomock.Any(), "ws-1", entity.ChatProviderSlack, entity.ChatFailing, "invalid_auth", gomock.Any()).Return(nil)

	if res := h.srv.Send(context.Background(), slackTarget(), entity.AlertPage{}); res.Delivered {
		t.Fatalf("expected an undelivered result, got %+v", res)
	}
}

func TestChatDeliverySuccessDoesNotWriteHealthWhenAlreadyHealthy(t *testing.T) {
	h := newHarness(t)
	h.expectChatSend(entity.ChatHealthy, entity.ChatSendResult{
		DMChannelID: "d1", Result: entity.NotifyResult{Delivered: true, MessageID: "m1"},
	})
	// no SetHealth expected: the desired state already matches conn.Health.

	if res := h.srv.Send(context.Background(), slackTarget(), entity.AlertPage{}); !res.Delivered {
		t.Fatalf("expected delivered, got %+v", res)
	}
}

func TestEmailCarriesAckResolveLinks(t *testing.T) {
	h := newHarness(t)
	var got entity.AlertPage
	h.mailer.EXPECT().SendPage(gomock.Any(), "on-call@acme.test", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, page entity.AlertPage) error { got = page; return nil })

	h.srv.Send(context.Background(), entity.NotifyTarget{
		Channel: entity.ChannelTypeEmail, Detail: "on-call@acme.test", AckToken: "atk", ResolveToken: "rtk",
	}, entity.AlertPage{})

	if got.AckURL != "https://opsy.test/v1/act/atk" || got.ResolveURL != "https://opsy.test/v1/act/rtk" {
		t.Fatalf("email ack/resolve URLs = %q / %q", got.AckURL, got.ResolveURL)
	}
}

func TestNtfyCarriesAckResolveActions(t *testing.T) {
	h := newHarness(t)
	var got entity.NtfyMessage
	h.ntfy.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msg entity.NtfyMessage) (entity.NotifyResult, error) {
			got = msg
			return entity.NotifyResult{Delivered: true, MessageID: "m1"}, nil
		})

	h.srv.Send(context.Background(), entity.NotifyTarget{
		Channel: entity.ChannelTypeNtfy, Detail: "ntfy.sh/pages-x", Secret: "tok", AckToken: "atk", ResolveToken: "rtk",
	}, entity.AlertPage{})

	if got.AckURL != "https://opsy.test/v1/act/atk" || got.ResolveURL != "https://opsy.test/v1/act/rtk" {
		t.Fatalf("ntfy ack/resolve URLs = %q / %q", got.AckURL, got.ResolveURL)
	}
}

func TestChatDeliverySuccessRecoversFailingHealth(t *testing.T) {
	h := newHarness(t)
	h.expectChatSend(entity.ChatFailing, entity.ChatSendResult{
		DMChannelID: "d1", Result: entity.NotifyResult{Delivered: true, MessageID: "m1"},
	})
	h.conns.EXPECT().SetHealth(gomock.Any(), "ws-1", entity.ChatProviderSlack, entity.ChatHealthy, "", gomock.Any()).Return(nil)

	if res := h.srv.Send(context.Background(), slackTarget(), entity.AlertPage{}); !res.Delivered {
		t.Fatalf("expected delivered, got %+v", res)
	}
}
