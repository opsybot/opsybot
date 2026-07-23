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

const (
	ChannelDetailMaxLength = 200
	ChannelLabelMaxLength  = 60
	ChannelVerifyTTL       = 15 * time.Minute
	ChannelVerifyEmailTTL  = 24 * time.Hour
	ChannelVerifyMaxTries  = 5
	ChannelVerifyCodeMax   = 1000000
)

type ChannelVerifyMethod string

const (
	ChannelVerifyEmail    ChannelVerifyMethod = "email"
	ChannelVerifyNtfy     ChannelVerifyMethod = "ntfy"
	ChannelVerifyWebhook  ChannelVerifyMethod = "webhook"
	ChannelVerifyTelegram ChannelVerifyMethod = "telegram"
	ChannelVerifyChat     ChannelVerifyMethod = "chat"
)

type Channel struct {
	ID        string
	UserID    string
	Type      ChannelType
	Detail    string
	Label     string
	Secret    string
	Verified  bool
	CreatedAt time.Time
}

type NewChannel struct {
	Type   ChannelType
	Detail string
	Label  string
	Secret string
}

type ChannelVerification struct {
	Method    ChannelVerifyMethod
	Token     string
	Code      string
	DeepLink  string
	Detail    string
	ExpiresAt time.Time
}

type ChannelVerifyRecord struct {
	ChannelID string
	UserID    string
	Method    ChannelVerifyMethod
	TokenHash string
	CodeHash  string
	Nonce     string
	ExpiresAt time.Time
}

type ChannelVerifyClaim struct {
	ChannelID string
	UserID    string
}

var (
	ErrChannelNotFound          = errors.New("channel not found")
	ErrChannelDuplicate         = errors.New("channel duplicate")
	ErrChannelVerifyExpired     = errors.New("channel verification expired")
	ErrChannelVerifyInvalid     = errors.New("channel verification invalid")
	ErrChannelSecretUnavailable = errors.New("channel secret storage unavailable")
	ErrChannelNotVerified       = errors.New("channel not verified")
)

func (t ChannelType) VerifyMethod() ChannelVerifyMethod {
	switch t {
	case ChannelTypeEmail:
		return ChannelVerifyEmail
	case ChannelTypeNtfy:
		return ChannelVerifyNtfy
	case ChannelTypeWebhook:
		return ChannelVerifyWebhook
	case ChannelTypeTelegram:
		return ChannelVerifyTelegram
	default:
		return ChannelVerifyChat
	}
}

func (t ChannelType) SelfServe() bool {
	switch t {
	case ChannelTypeEmail, ChannelTypeNtfy, ChannelTypeWebhook:
		return true
	default:
		return false
	}
}

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
