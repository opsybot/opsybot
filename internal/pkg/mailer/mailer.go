package mailer

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/wneessen/go-mail"

	"github.com/opsybot/opsybot/internal/config"
)

//go:embed templates/*.txt.tmpl
var templateFS embed.FS

var ErrDisabled = errors.New("mailer disabled: no mailer.host configured")

type Client struct {
	client   *mail.Client
	from     string
	fromName string
	tmpl     *template.Template
	enabled  bool
}

type InviteData struct {
	InviterName   string
	WorkspaceName string
	AcceptURL     string
}

type ResetData struct {
	ResetURL string
}

type PageData struct {
	Severity   string
	Service    string
	Title      string
	StartedAt  string
	PolicySlug string
	Level      int
	AlertURL   string
	AckURL     string
	ResolveURL string
}

func New(cfg config.Mailer) (Client, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.txt.tmpl")
	if err != nil {
		return Client{}, fmt.Errorf("parse mail templates: %w", err)
	}
	if cfg.Host == "" {
		return Client{from: cfg.From, fromName: cfg.FromName, tmpl: tmpl, enabled: false}, nil
	}

	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithTimeout(cfg.Timeout),
		mail.WithTLSPolicy(tlsPolicy(cfg.Encryption)),
	}
	if cfg.Username != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(cfg.Username), mail.WithPassword(cfg.Password))
	}
	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return Client{}, fmt.Errorf("new mail client: %w", err)
	}
	return Client{client: client, from: cfg.From, fromName: cfg.FromName, tmpl: tmpl, enabled: true}, nil
}

func tlsPolicy(encryption string) mail.TLSPolicy {
	switch encryption {
	case "tls":
		return mail.TLSMandatory
	case "none":
		return mail.NoTLS
	default:
		return mail.TLSOpportunistic
	}
}

func (c Client) Enabled() bool { return c.enabled }

func (c Client) SendInvite(ctx context.Context, to string, data InviteData) error {
	return c.send(ctx, to, "You've been invited to Opsybot", "invite.txt.tmpl", data)
}

func (c Client) SendPasswordReset(ctx context.Context, to string, data ResetData) error {
	return c.send(ctx, to, "Reset your Opsybot password", "reset.txt.tmpl", data)
}

func (c Client) SendPage(ctx context.Context, to string, data PageData) error {
	subject := "[" + strings.ToUpper(data.Severity) + "] " + data.Service + ": " + data.Title
	return c.send(ctx, to, subject, "page.txt.tmpl", data)
}

func (c Client) SendText(ctx context.Context, to, subject, body string) error {
	if !c.enabled {
		return ErrDisabled
	}
	msg := mail.NewMsg()
	if err := msg.FromFormat(c.fromName, c.from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)
	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func (c Client) send(ctx context.Context, to, subject, tmpl string, data any) error {
	if !c.enabled {
		return ErrDisabled
	}
	var body strings.Builder
	if err := c.tmpl.ExecuteTemplate(&body, tmpl, data); err != nil {
		return fmt.Errorf("render mail %s: %w", tmpl, err)
	}
	msg := mail.NewMsg()
	if err := msg.FromFormat(c.fromName, c.from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body.String())
	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
