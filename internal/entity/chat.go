package entity

import (
	"errors"
	"time"
)

type ChatProvider string

const (
	ChatProviderSlack    ChatProvider = "slack"
	ChatProviderTeams    ChatProvider = "teams"
	ChatProviderDiscord  ChatProvider = "discord"
	ChatProviderTelegram ChatProvider = "telegram"
)

type ChatHealth string

const (
	ChatHealthy ChatHealth = "healthy"
	ChatFailing ChatHealth = "failing"
)

type ChatConnection struct {
	ID               string
	WorkspaceID      string
	Provider         ChatProvider
	ExternalID       string
	ExternalName     string
	BotUserID        string
	HasBotToken      bool
	AppID            string
	TenantID         string
	Scopes           []string
	Enabled          bool
	Health           ChatHealth
	HealthNote       string
	HealthCheckedAt  time.Time
	NamingPattern    string
	AnnounceChannel  string
	ArchiveOnResolve bool
	ConnectedBy      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Linked           bool
	LinkedHandle     string
	LinkedVerified   bool
}

type ChatConnectionInput struct {
	Provider         ChatProvider
	ExternalID       string
	ExternalName     string
	BotUserID        string
	BotToken         string
	AppID            string
	Scopes           []string
	NamingPattern    string
	AnnounceChannel  string
	ArchiveOnResolve bool
	ConnectedBy      string
}

type ChatConnectInput struct {
	Provider   ChatProvider
	BotToken   string
	ExternalID string
}

type ChatIdentity struct {
	ID             string
	ConnectionID   string
	UserID         string
	ProviderUserID string
	ProviderHandle string
	DMChannelID    string
	ResolvedBy     string
	Verified       bool
}

type ChatCallback struct {
	Provider  ChatProvider
	Body      []byte
	Signature string
	Timestamp string
	IP        string
}

type ChatDelivery struct {
	Provider       ChatProvider
	BotToken       string
	ProviderUserID string
	DMChannelID    string
	Page           AlertPage
	AckToken       string
	ResolveToken   string
	BaseURL        string
}

type ChatSendResult struct {
	DMChannelID string
	Result      NotifyResult
}

type ChatValidation struct {
	ExternalID   string
	ExternalName string
	BotUserID    string
}

type ChatOAuthPurpose string

const (
	ChatOAuthInstall  ChatOAuthPurpose = "install"
	ChatOAuthIdentity ChatOAuthPurpose = "identity"
	ChatOAuthLink     ChatOAuthPurpose = "link"
)

func TelegramWebhookSecret(botToken string) string {
	return HashToken(botToken)
}

type ChatOAuthState struct {
	Provider      ChatProvider
	Purpose       ChatOAuthPurpose
	WorkspaceID   string
	WorkspaceSlug string
	UserID        string
	ConnectionID  string
	TeamID        string
}

type ChatOAuthResult struct {
	ExternalID   string
	ExternalName string
	BotUserID    string
	BotToken     string
	Scopes       []string
}

type ChatIdentityResult struct {
	ProviderUserID string
	TeamID         string
	Handle         string
	Email          string
}

type ChatUser struct {
	ProviderUserID string
	Handle         string
}

type InteractionResponse struct {
	Status      int
	ContentType string
	Body        []byte
}

type ActionKind string

const (
	ActionKindAck     ActionKind = "ack"
	ActionKindResolve ActionKind = "resolve"
)

const (
	ActionTokenLength    = 32
	ChatBotTokenMax      = 400
	ChatOAuthStateLength = 32
	ChatOAuthStateTTL    = 10 * time.Minute
)

var SlackOAuthScopes = []string{"chat:write", "im:write", "users:read", "users:read.email"}

var SlackOIDCScopes = []string{"openid", "profile", "email"}

var DiscordBotScopes = []string{"bot"}

var DiscordIdentityScopes = []string{"identify"}

const DiscordBotPermissions = "68624"

const DefaultAnnounceChannel = "#incidents"

type AlertAction struct {
	Token  string
	Action ActionKind
}

type AlertActionRecord struct {
	WorkspaceID string
	AlertID     string
	UserID      string
	ChannelID   string
	Action      ActionKind
	TokenHash   string
	ExpiresAt   time.Time
}

type ActionClaim struct {
	WorkspaceID string
	AlertID     string
	UserID      string
	Action      ActionKind
}

type ActionOutcome struct {
	Action     ActionKind
	AlertTitle string
	Actor      string
	At         time.Time
}

var (
	ErrChatNotConnected          = errors.New("chat provider not connected")
	ErrChatConnectionInvalid     = errors.New("chat connection invalid")
	ErrChatProviderNotConfigured = errors.New("chat provider not configured")
	ErrChatSecretUnavailable     = errors.New("chat secret storage unavailable")
	ErrChatSignatureInvalid      = errors.New("chat signature invalid")
	ErrChatOAuthUnsupported      = errors.New("chat provider does not support oauth")
	ErrChatOAuthStateInvalid     = errors.New("chat oauth state invalid")
	ErrChatOAuthExchange         = errors.New("chat oauth exchange failed")
	ErrActionTokenInvalid        = errors.New("action token invalid")
	ErrActionProviderUnknown     = errors.New("action provider unknown")
)

func (p ChatProvider) Valid() bool {
	switch p {
	case ChatProviderSlack, ChatProviderTeams, ChatProviderDiscord, ChatProviderTelegram:
		return true
	default:
		return false
	}
}
