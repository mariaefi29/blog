package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPPort    = 8080
	defaultHTTPTimeout = 30 * time.Second
	defaultStaticDir   = "public"
	defaultDatabase    = "blog"
	defaultSMTPHost    = "smtp.mail.ru"
	defaultSMTPPort    = 465
)

type Config struct {
	HTTP      HTTPConfig
	Mongo     MongoConfig
	SMTP      SMTPConfig
	Recaptcha RecaptchaConfig
	Analytics AnalyticsConfig
}

type HTTPConfig struct {
	Port      int
	Timeout   time.Duration
	StaticDir string
}

type MongoConfig struct {
	ConnectionString string
	Database         string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Email    string
	Password string
}

type RecaptchaConfig struct {
	Secret string
}

type AnalyticsConfig struct {
	GoogleAnalyticsMeasurementID string
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Port:      defaultHTTPPort,
			Timeout:   defaultHTTPTimeout,
			StaticDir: getenv("STATIC_DIR", defaultStaticDir),
		},
		Mongo: MongoConfig{
			ConnectionString: os.Getenv("DB_CONNECTION_STRING"),
			Database:         getenv("DB_NAME", defaultDatabase),
		},
		SMTP: SMTPConfig{
			Host:     getenv("SMTP_HOST", defaultSMTPHost),
			Port:     defaultSMTPPort,
			Email:    os.Getenv("SMTP_EMAIL"),
			Password: os.Getenv("SMTP_PASSWORD"),
		},
		Recaptcha: RecaptchaConfig{
			Secret: os.Getenv("RECAPTCHA_SECRET"),
		},
		Analytics: AnalyticsConfig{
			GoogleAnalyticsMeasurementID: os.Getenv("GOOGLE_ANALYTICS_MEASUREMENT_ID"),
		},
	}

	port, err := intFromEnv("PORT", cfg.HTTP.Port)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.Port = port

	smtpPort, err := intFromEnv("SMTP_PORT", cfg.SMTP.Port)
	if err != nil {
		return Config{}, err
	}
	cfg.SMTP.Port = smtpPort

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", key, value, err)
	}

	return parsed, nil
}
