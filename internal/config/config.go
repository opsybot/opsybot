package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment     string        `mapstructure:"environment"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Log             Log           `mapstructure:"log"`
	OTel            OTel          `mapstructure:"otel"`
	HTTP            HTTP          `mapstructure:"http"`
	Postgres        Postgres      `mapstructure:"postgres"`
	Valkey          Valkey        `mapstructure:"valkey"`
	Casbin          Casbin        `mapstructure:"casbin"`
	Auth            Auth          `mapstructure:"auth"`
	Mailer          Mailer        `mapstructure:"mailer"`
	Ingest          Ingest        `mapstructure:"ingest"`
	Cron            Cron          `mapstructure:"cron"`
	Webhook         Webhook       `mapstructure:"webhook"`
	Ntfy            Ntfy          `mapstructure:"ntfy"`
	Slack           Slack         `mapstructure:"slack"`
	Discord         Discord       `mapstructure:"discord"`
	Telegram        Telegram      `mapstructure:"telegram"`
	Teams           Teams         `mapstructure:"teams"`
	Chat            Chat          `mapstructure:"chat"`
}

type Webhook struct {
	Timeout   time.Duration `mapstructure:"timeout"`
	UserAgent string        `mapstructure:"user_agent"`
}

type Ntfy struct {
	DefaultServer string        `mapstructure:"default_server"`
	Timeout       time.Duration `mapstructure:"timeout"`
	UserAgent     string        `mapstructure:"user_agent"`
}

type Slack struct {
	ClientID      string        `mapstructure:"client_id"`
	ClientSecret  string        `mapstructure:"client_secret"`
	SigningSecret string        `mapstructure:"signing_secret"`
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	UserAgent     string        `mapstructure:"user_agent"`
}

type Discord struct {
	ApplicationID string        `mapstructure:"application_id"`
	ClientID      string        `mapstructure:"client_id"`
	ClientSecret  string        `mapstructure:"client_secret"`
	PublicKey     string        `mapstructure:"public_key"`
	BotToken      string        `mapstructure:"bot_token"`
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	UserAgent     string        `mapstructure:"user_agent"`
}

type Telegram struct {
	BotToken  string        `mapstructure:"bot_token"`
	BotName   string        `mapstructure:"bot_name"`
	BaseURL   string        `mapstructure:"base_url"`
	Timeout   time.Duration `mapstructure:"timeout"`
	UserAgent string        `mapstructure:"user_agent"`
}

type Teams struct {
	AppID        string        `mapstructure:"app_id"`
	AppSecret    string        `mapstructure:"app_secret"`
	TenantID     string        `mapstructure:"tenant_id"`
	CatalogAppID string        `mapstructure:"catalog_app_id"`
	GraphBaseURL string        `mapstructure:"graph_base_url"`
	BotBaseURL   string        `mapstructure:"bot_base_url"`
	LoginBaseURL string        `mapstructure:"login_base_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
	UserAgent    string        `mapstructure:"user_agent"`
}

type Chat struct {
	InteractionSkew time.Duration `mapstructure:"interaction_skew"`
	ActionTokenTTL  time.Duration `mapstructure:"action_token_ttl"`
}

type Cron struct {
	HeartbeatSweep    time.Duration `mapstructure:"heartbeat_sweep"`
	AlertAutoResolve  time.Duration `mapstructure:"alert_autoresolve"`
	IngestRetention   string        `mapstructure:"ingest_retention"`
	EscalationSweep   time.Duration `mapstructure:"escalation_sweep"`
	NotificationSweep time.Duration `mapstructure:"notification_sweep"`
	JobTimeout        time.Duration `mapstructure:"job_timeout"`
	LockExpiry        time.Duration `mapstructure:"lock_expiry"`
	StopTimeout       time.Duration `mapstructure:"stop_timeout"`
}

type Ingest struct {
	BaseURL          string        `mapstructure:"base_url"`
	MaxBodyBytes     int64         `mapstructure:"max_body_bytes"`
	MaxConcurrent    int           `mapstructure:"max_concurrent"`
	RatePerMin       int           `mapstructure:"rate_per_min"`
	FailureRetention time.Duration `mapstructure:"failure_retention"`
}

type Mailer struct {
	Host       string        `mapstructure:"host"`
	Port       int           `mapstructure:"port"`
	Username   string        `mapstructure:"username"`
	Password   string        `mapstructure:"password"`
	Encryption string        `mapstructure:"encryption"`
	From       string        `mapstructure:"from"`
	FromName   string        `mapstructure:"from_name"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

type Auth struct {
	BaseURL                string        `mapstructure:"base_url"`
	SecretKey              string        `mapstructure:"secret_key"`
	SecretKeyPrevious      string        `mapstructure:"secret_key_previous"`
	CookieName             string        `mapstructure:"cookie_name"`
	CookieSecure           bool          `mapstructure:"cookie_secure"`
	SessionIdleTTL         time.Duration `mapstructure:"session_idle_ttl"`
	SessionAbsoluteTTL     time.Duration `mapstructure:"session_absolute_ttl"`
	SessionBrowserTTL      time.Duration `mapstructure:"session_browser_ttl"`
	SessionTouchWindow     time.Duration `mapstructure:"session_touch_window"`
	TrustProxyHeaders      bool          `mapstructure:"trust_proxy_headers"`
	InviteTTL              time.Duration `mapstructure:"invite_ttl"`
	RateLoginPerMin        int           `mapstructure:"rate_login_per_min"`
	RateSignupPerHour      int           `mapstructure:"rate_signup_per_hour"`
	RateSlugCheckPerMin    int           `mapstructure:"rate_slug_check_per_min"`
	RateResetPerHour       int           `mapstructure:"rate_reset_per_hour"`
	RateSSOPerMin          int           `mapstructure:"rate_sso_per_min"`
	RateNotifyPerMin       int           `mapstructure:"rate_notify_per_min"`
	RateChannelTestPerHour int           `mapstructure:"rate_channel_test_per_hour"`
}

type Environment string

type OTel struct {
	Endpoint       string        `mapstructure:"endpoint"`
	Insecure       bool          `mapstructure:"insecure"`
	ServiceName    string        `mapstructure:"service_name"`
	SampleRatio    float64       `mapstructure:"sample_ratio"`
	MetricInterval time.Duration `mapstructure:"metric_interval"`
	ExportTimeout  time.Duration `mapstructure:"export_timeout"`
}

type HTTP struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
}

type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type Postgres struct {
	URL             string        `mapstructure:"url"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
}

type Valkey struct {
	Addrs        []string      `mapstructure:"addrs"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
}

type Casbin struct {
	TableName string `mapstructure:"table_name"`
	Channel   string `mapstructure:"channel"`
}

func NewPostgres(cfg Config) Postgres {
	return cfg.Postgres
}

func NewLog(cfg Config) Log {
	return cfg.Log
}

func NewHTTP(cfg Config) HTTP {
	return cfg.HTTP
}

func NewOTel(cfg Config) OTel {
	return cfg.OTel
}

func NewAuth(cfg Config) Auth {
	return cfg.Auth
}

func NewIngest(cfg Config) Ingest {
	return cfg.Ingest
}

func NewCron(cfg Config) Cron {
	return cfg.Cron
}

func NewWebhook(cfg Config) Webhook {
	return cfg.Webhook
}

func NewNtfy(cfg Config) Ntfy {
	return cfg.Ntfy
}

func NewSlack(cfg Config) Slack {
	return cfg.Slack
}

func NewDiscord(cfg Config) Discord {
	return cfg.Discord
}

func NewTelegram(cfg Config) Telegram {
	return cfg.Telegram
}

func NewTeams(cfg Config) Teams {
	return cfg.Teams
}

func NewChat(cfg Config) Chat {
	return cfg.Chat
}

func NewMailer(cfg Config) Mailer {
	return cfg.Mailer
}

func NewEnvironment(cfg Config) Environment {
	return Environment(cfg.Environment)
}

func NewValkey(cfg Config) Valkey {
	return cfg.Valkey
}

func NewCasbin(cfg Config) Casbin {
	return cfg.Casbin
}

func New(cfgFile string) (Config, error) {
	v := viper.New()

	v.SetDefault("environment", "production")
	v.SetDefault("shutdown_timeout", "15s")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("otel.endpoint", "")
	v.SetDefault("otel.insecure", true)
	v.SetDefault("otel.service_name", "opsybot")
	v.SetDefault("otel.sample_ratio", 1.0)
	v.SetDefault("otel.metric_interval", "60s")
	v.SetDefault("otel.export_timeout", "10s")
	v.SetDefault("http.host", "")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_header_timeout", "5s")
	v.SetDefault("http.read_timeout", "30s")
	v.SetDefault("http.write_timeout", "30s")
	v.SetDefault("http.idle_timeout", "120s")
	v.SetDefault("postgres.url", "")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "opsybot")
	v.SetDefault("postgres.password", "")
	v.SetDefault("postgres.database", "opsybot")
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("postgres.max_open_conns", 25)
	v.SetDefault("postgres.max_idle_conns", 25)
	v.SetDefault("postgres.conn_max_lifetime", "5m")
	v.SetDefault("postgres.conn_max_idle_time", "5m")
	v.SetDefault("postgres.connect_timeout", "5s")
	v.SetDefault("valkey.addrs", []string{"localhost:6379"})
	v.SetDefault("valkey.username", "")
	v.SetDefault("valkey.password", "")
	v.SetDefault("valkey.db", 0)
	v.SetDefault("valkey.dial_timeout", "5s")
	v.SetDefault("valkey.read_timeout", "3s")
	v.SetDefault("valkey.write_timeout", "3s")
	v.SetDefault("valkey.pool_size", 10)
	v.SetDefault("casbin.table_name", "casbin_rule")
	v.SetDefault("casbin.channel", "/casbin")
	v.SetDefault("auth.base_url", "http://localhost:8080")
	v.SetDefault("auth.secret_key", "")
	v.SetDefault("auth.secret_key_previous", "")
	v.SetDefault("auth.cookie_name", "opsybot_session")
	v.SetDefault("auth.cookie_secure", true)
	v.SetDefault("auth.session_idle_ttl", "72h")
	v.SetDefault("auth.session_absolute_ttl", "720h")
	v.SetDefault("auth.session_browser_ttl", "24h")
	v.SetDefault("auth.session_touch_window", "5m")
	v.SetDefault("auth.trust_proxy_headers", false)
	v.SetDefault("auth.invite_ttl", "336h")
	v.SetDefault("auth.rate_login_per_min", 10)
	v.SetDefault("auth.rate_signup_per_hour", 5)
	v.SetDefault("auth.rate_slug_check_per_min", 60)
	v.SetDefault("auth.rate_reset_per_hour", 5)
	v.SetDefault("auth.rate_sso_per_min", 20)
	v.SetDefault("auth.rate_notify_per_min", 60)
	v.SetDefault("auth.rate_channel_test_per_hour", 10)
	v.SetDefault("ingest.base_url", "")
	v.SetDefault("ingest.max_body_bytes", 1048576)
	v.SetDefault("ingest.max_concurrent", 64)
	v.SetDefault("ingest.rate_per_min", 600)
	v.SetDefault("ingest.failure_retention", "168h")
	v.SetDefault("cron.heartbeat_sweep", "30s")
	v.SetDefault("cron.alert_autoresolve", "5m")
	v.SetDefault("cron.ingest_retention", "17 3 * * *")
	v.SetDefault("cron.job_timeout", "2m")
	v.SetDefault("cron.lock_expiry", "60s")
	v.SetDefault("cron.stop_timeout", "30s")
	v.SetDefault("cron.escalation_sweep", "10s")
	v.SetDefault("cron.notification_sweep", "5s")
	v.SetDefault("webhook.timeout", "10s")
	v.SetDefault("webhook.user_agent", "opsybot")
	v.SetDefault("ntfy.default_server", "https://ntfy.sh")
	v.SetDefault("ntfy.timeout", "10s")
	v.SetDefault("ntfy.user_agent", "opsybot")
	v.SetDefault("slack.client_id", "")
	v.SetDefault("slack.client_secret", "")
	v.SetDefault("slack.signing_secret", "")
	v.SetDefault("slack.bot_token", "")
	v.SetDefault("slack.base_url", "https://slack.com/api")
	v.SetDefault("slack.timeout", "10s")
	v.SetDefault("slack.user_agent", "opsybot")
	v.SetDefault("discord.application_id", "")
	v.SetDefault("discord.client_id", "")
	v.SetDefault("discord.client_secret", "")
	v.SetDefault("discord.public_key", "")
	v.SetDefault("discord.bot_token", "")
	v.SetDefault("discord.base_url", "https://discord.com/api/v10")
	v.SetDefault("discord.timeout", "10s")
	v.SetDefault("discord.user_agent", "opsybot")
	v.SetDefault("telegram.bot_token", "")
	v.SetDefault("telegram.bot_name", "")
	v.SetDefault("telegram.base_url", "https://api.telegram.org")
	v.SetDefault("telegram.timeout", "10s")
	v.SetDefault("telegram.user_agent", "opsybot")
	v.SetDefault("chat.interaction_skew", "5m")
	v.SetDefault("chat.action_token_ttl", "24h")
	v.SetDefault("mailer.host", "")
	v.SetDefault("mailer.port", 587)
	v.SetDefault("mailer.username", "")
	v.SetDefault("mailer.password", "")
	v.SetDefault("mailer.encryption", "starttls")
	v.SetDefault("mailer.from", "opsybot@localhost")
	v.SetDefault("mailer.from_name", "Opsybot")
	v.SetDefault("mailer.timeout", "10s")

	v.SetEnvPrefix("OPSYBOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config file %s: %w", cfgFile, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
