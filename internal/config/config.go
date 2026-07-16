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
	Postgres        Postgres      `mapstructure:"postgres"`
	Valkey          Valkey        `mapstructure:"valkey"`
	Casbin          Casbin        `mapstructure:"casbin"`
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
