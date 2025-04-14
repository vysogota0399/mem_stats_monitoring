package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	err := os.Setenv("POLL_INTERVAL", "1")
	assert.NoError(t, err)
	err = os.Setenv("REPORT_INTERVAL", "2")
	assert.NoError(t, err)
	err = os.Setenv("ADDRESS", "localhost")
	assert.NoError(t, err)
	err = os.Setenv("LOG_LEVEL", "-1")
	assert.NoError(t, err)
	err = os.Setenv("KEY", "secret")
	assert.NoError(t, err)
	err = os.Setenv("RATE_LIMIT", "10")
	assert.NoError(t, err)
	err = os.Setenv("CRYPTO_KEY", "")
	assert.NoError(t, err)

	cfg, err := NewConfig(nil)
	assert.NoError(t, err)

	assert.Equal(t, cfg.PollInterval, time.Duration(1)*time.Second)
	assert.Equal(t, cfg.ReportInterval, time.Duration(2)*time.Second)
	assert.Equal(t, cfg.ServerURL, "http://localhost")
	assert.Equal(t, cfg.LogLevel, int64(-1))
	assert.Equal(t, cfg.Key, "secret")
	assert.Equal(t, cfg.RateLimit, 10)
}
