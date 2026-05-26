package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != 8080 {
		t.Fatalf("HTTP port = %d, want 8080", cfg.HTTP.Port)
	}
	if cfg.HTTP.Timeout != 30*time.Second {
		t.Fatalf("HTTP timeout = %s, want 30s", cfg.HTTP.Timeout)
	}
	if cfg.HTTP.StaticDir != "public" {
		t.Fatalf("static dir = %q, want public", cfg.HTTP.StaticDir)
	}
	if cfg.Mongo.Database != "blog" {
		t.Fatalf("mongo database = %q, want blog", cfg.Mongo.Database)
	}
	if cfg.SMTP.Host != "smtp.mail.ru" {
		t.Fatalf("smtp host = %q, want smtp.mail.ru", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != 465 {
		t.Fatalf("smtp port = %d, want 465", cfg.SMTP.Port)
	}
}

func TestLoadOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("STATIC_DIR", "assets")
	t.Setenv("DB_CONNECTION_STRING", "mongodb://localhost:27017")
	t.Setenv("DB_NAME", "blog_test")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_EMAIL", "admin@example.com")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("RECAPTCHA_SECRET", "recaptcha-secret")
	t.Setenv("GOOGLE_ANALYTICS_MEASUREMENT_ID", "G-123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != 9090 {
		t.Fatalf("HTTP port = %d, want 9090", cfg.HTTP.Port)
	}
	if cfg.HTTP.StaticDir != "assets" {
		t.Fatalf("static dir = %q, want assets", cfg.HTTP.StaticDir)
	}
	if cfg.Mongo.ConnectionString != "mongodb://localhost:27017" {
		t.Fatalf("mongo connection string = %q", cfg.Mongo.ConnectionString)
	}
	if cfg.Mongo.Database != "blog_test" {
		t.Fatalf("mongo database = %q, want blog_test", cfg.Mongo.Database)
	}
	if cfg.SMTP.Host != "smtp.example.com" {
		t.Fatalf("smtp host = %q, want smtp.example.com", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != 2525 {
		t.Fatalf("smtp port = %d, want 2525", cfg.SMTP.Port)
	}
	if cfg.SMTP.Email != "admin@example.com" {
		t.Fatalf("smtp email = %q, want admin@example.com", cfg.SMTP.Email)
	}
	if cfg.SMTP.Password != "secret" {
		t.Fatalf("smtp password = %q, want secret", cfg.SMTP.Password)
	}
	if cfg.Recaptcha.Secret != "recaptcha-secret" {
		t.Fatalf("recaptcha secret = %q, want recaptcha-secret", cfg.Recaptcha.Secret)
	}
	if cfg.Analytics.GoogleAnalyticsMeasurementID != "G-123" {
		t.Fatalf("google analytics id = %q, want G-123", cfg.Analytics.GoogleAnalyticsMeasurementID)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid PORT value") {
		t.Fatalf("Load() error = %q, want invalid PORT value", err)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"PORT",
		"STATIC_DIR",
		"DB_CONNECTION_STRING",
		"DB_NAME",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_EMAIL",
		"SMTP_PASSWORD",
		"RECAPTCHA_SECRET",
		"GOOGLE_ANALYTICS_MEASUREMENT_ID",
	} {
		t.Setenv(key, "")
	}
}
