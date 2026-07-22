package entity

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ChannelType string

const (
	ChannelTypeSlack    ChannelType = "slack"
	ChannelTypeTeams    ChannelType = "teams"
	ChannelTypeDiscord  ChannelType = "discord"
	ChannelTypeTelegram ChannelType = "telegram"
	ChannelTypeNtfy     ChannelType = "ntfy"
	ChannelTypeEmail    ChannelType = "email"
	ChannelTypeWebhook  ChannelType = "webhook"
)

const ChannelDetailMaxLength = 200

type Channel struct {
	ID        string
	UserID    string
	Type      ChannelType
	Detail    string
	Verified  bool
	CreatedAt time.Time
}

type NewChannel struct {
	Type   ChannelType
	Detail string
}

var (
	ErrChannelNotFound  = errors.New("channel not found")
	ErrChannelDuplicate = errors.New("channel duplicate")
)

func (t ChannelType) Valid() bool {
	switch t {
	case ChannelTypeSlack, ChannelTypeTeams, ChannelTypeDiscord, ChannelTypeTelegram,
		ChannelTypeNtfy, ChannelTypeEmail, ChannelTypeWebhook:
		return true
	default:
		return false
	}
}

func (t ChannelType) EventKind() AlertEventKind {
	switch t {
	case ChannelTypeSlack, ChannelTypeTeams, ChannelTypeDiscord, ChannelTypeTelegram:
		return AlertEventChat
	case ChannelTypeNtfy:
		return AlertEventPush
	default:
		return AlertEventNotified
	}
}

func (n NewChannel) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Type, validation.By(channelTypeField)),
		validation.Field(&n.Detail, validation.By(func(value any) error {
			return channelDetailFor(n.Type, value)
		})),
	)
}
