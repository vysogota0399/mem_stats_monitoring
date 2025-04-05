package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	os.Setenv("POLL_INTERVAL", "1")
	os.Setenv("REPORT_INTERVAL", "2")
	os.Setenv("ADDRESS", "localhost")
	os.Setenv("LOG_LEVEL", "-1")
	os.Setenv("KEY", "secret")
	os.Setenv("RATE_LIMIT", "10")

	cfg, err := NewConfig()
	assert.NoError(t, err)

	assert.Equal(t, cfg.PollInterval, time.Duration(1)*time.Second)
	assert.Equal(t, cfg.ReportInterval, time.Duration(2)*time.Second)
	assert.Equal(t, cfg.ServerURL, "http://localhost")
	assert.Equal(t, cfg.LogLevel, int64(-1))
	assert.Equal(t, cfg.Key, "secret")
	assert.Equal(t, cfg.RateLimit, 10)
}
