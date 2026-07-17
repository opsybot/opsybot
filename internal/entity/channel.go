package entity

import (
	"errors"
	"net/url"
	"strings"
	"time"
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
	ErrChannelInvalid   = errors.New("channel invalid")
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

func (n NewChannel) Validate() error {
	if !n.Type.Valid() {
		return ErrChannelInvalid
	}
	detail := strings.TrimSpace(n.Detail)
	if detail == "" || len(detail) > ChannelDetailMaxLength {
		return ErrChannelInvalid
	}
	switch n.Type {
	case ChannelTypeEmail:
		if err := ValidateEmail(detail); err != nil {
			return ErrChannelInvalid
		}
	case ChannelTypeWebhook, ChannelTypeNtfy:
		u, err := url.ParseRequestURI(detail)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return ErrChannelInvalid
		}
	}
	return nil
}
