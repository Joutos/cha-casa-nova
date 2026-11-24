// Package config carrega a configuração a partir de variáveis de ambiente.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	DateLabel    string
	Time         string
	AddressLine1 string
	AddressLine2 string
	MapsQuery    string
	Signature    string
}

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (s SMTP) Enabled() bool { return s.Host != "" && s.From != "" }

type Config struct {
	Env            string
	Port           string
	DatabasePath   string
	AllowedOrigins []string
	TrustedProxies []string
	SessionTTL     time.Duration
	SessionSecret  []byte
	CookieSecure   bool
	LogLevel       slog.Level
	PublicBaseURL  string

	LoginRatePerMinute  int
	PublicRatePerMinute int
	MaxMessagesPerGuest int

	BootstrapAdminName     string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	Event Event
	SMTP  SMTP
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func Load() (*Config, error) {
	c := &Config{
		Env:                 get("APP_ENV", "development"),
		Port:                get("PORT", "8080"),
		DatabasePath:        get("DATABASE_PATH", "./data/app.db"),
		AllowedOrigins:      list(get("ALLOWED_ORIGINS", "")),
		TrustedProxies:      list(get("TRUSTED_PROXIES", "127.0.0.1,::1,172.16.0.0/12,10.0.0.0/8,192.168.0.0/16")),
		PublicBaseURL:       strings.TrimRight(get("PUBLIC_BASE_URL", "http://localhost:8080"), "/"),
		LoginRatePerMinute:  getInt("LOGIN_RATE_PER_MINUTE", 5),
		PublicRatePerMinute: getInt("PUBLIC_RATE_PER_MINUTE", 10),
		MaxMessagesPerGuest: getInt("MAX_MESSAGES_PER_GUEST", 5),

		BootstrapAdminName:     get("ADMIN_NAME", "Administrador"),
		BootstrapAdminEmail:    get("ADMIN_EMAIL", ""),
		BootstrapAdminPassword: os.Getenv("ADMIN_PASSWORD"),

		Event: Event{
			DateLabel:    get("EVENT_DATE_LABEL", "23 de Janeiro de 2027"),
			Time:         get("EVENT_TIME", "11h00"),
			AddressLine1: get("EVENT_ADDRESS_LINE1", ""),
			AddressLine2: get("EVENT_ADDRESS_LINE2", ""),
			MapsQuery:    get("EVENT_MAPS_QUERY", ""),
			Signature:    get("EVENT_SIGNATURE", "Sofia e João"),
		},
		SMTP: SMTP{
			Host:     get("SMTP_HOST", ""),
			Port:     getInt("SMTP_PORT", 587),
			Username: get("SMTP_USERNAME", ""),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     get("SMTP_FROM", ""),
		},
	}

	ttl, err := time.ParseDuration(get("SESSION_TTL", "8h"))
	if err != nil || ttl < time.Minute {
		return nil, fmt.Errorf("SESSION_TTL inválido: %q", os.Getenv("SESSION_TTL"))
	}
	c.SessionTTL = ttl

	switch strings.ToLower(get("LOG_LEVEL", "info")) {
	case "debug":
		c.LogLevel = slog.LevelDebug
	case "warn":
		c.LogLevel = slog.LevelWarn
	case "error":
		c.LogLevel = slog.LevelError
	default:
		c.LogLevel = slog.LevelInfo
	}

	c.CookieSecure = getBool("COOKIE_SECURE", c.IsProduction())

	secret := os.Getenv("SESSION_SECRET")
	switch {
	case len(secret) >= 32:
		c.SessionSecret = []byte(secret)
	case c.IsProduction():
		return nil, errors.New("SESSION_SECRET é obrigatório em produção (mínimo de 32 caracteres)")
	default:
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		c.SessionSecret = []byte(hex.EncodeToString(buf))
		slog.Warn("SESSION_SECRET ausente: usando um segredo aleatório (sessões caem a cada reinício)")
	}

	if c.IsProduction() && !c.CookieSecure {
		slog.Warn("COOKIE_SECURE=false em produção: use apenas atrás de HTTPS")
	}
	return c, nil
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func getInt(key string, def int) int {
	if v, err := strconv.Atoi(get(key, "")); err == nil && v > 0 {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, err := strconv.ParseBool(get(key, "")); err == nil {
		return v
	}
	return def
}

func list(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
