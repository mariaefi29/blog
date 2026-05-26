package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	require.NoError(t, err)

	require.Equal(t, 8080, cfg.HTTP.Port)
	require.Equal(t, 30*time.Second, cfg.HTTP.Timeout)
	require.Equal(t, "public", cfg.HTTP.StaticDir)
	require.Equal(t, "blog", cfg.Mongo.Database)
	require.Equal(t, "smtp.mail.ru", cfg.SMTP.Host)
	require.Equal(t, 465, cfg.SMTP.Port)
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
	require.NoError(t, err)

	require.Equal(t, 9090, cfg.HTTP.Port)
	require.Equal(t, "assets", cfg.HTTP.StaticDir)
	require.Equal(t, "mongodb://localhost:27017", cfg.Mongo.ConnectionString)
	require.Equal(t, "blog_test", cfg.Mongo.Database)
	require.Equal(t, "smtp.example.com", cfg.SMTP.Host)
	require.Equal(t, 2525, cfg.SMTP.Port)
	require.Equal(t, "admin@example.com", cfg.SMTP.Email)
	require.Equal(t, "secret", cfg.SMTP.Password)
	require.Equal(t, "recaptcha-secret", cfg.Recaptcha.Secret)
	require.Equal(t, "G-123", cfg.Analytics.GoogleAnalyticsMeasurementID)
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "not-a-number")

	_, err := Load()
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid PORT value")
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
