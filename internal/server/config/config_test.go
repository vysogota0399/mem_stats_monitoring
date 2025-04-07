package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	err := os.Setenv("ADDRESS", "localhost")
	assert.NoError(t, err)
	err = os.Setenv("APP_ENV", "development")
	assert.NoError(t, err)
	err = os.Setenv("LOG_LEVEL", "0")
	assert.NoError(t, err)
	err = os.Setenv("STORE_INTERVAL", "10")
	assert.NoError(t, err)
	err = os.Setenv("FILE_STORAGE_PATH", "/tmp")
	assert.NoError(t, err)
	err = os.Setenv("RESTORE", "false")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_DSN", "pg@pg")
	assert.NoError(t, err)

	cfg, err := NewConfig()
	assert.NoError(t, err)

	assert.Equal(t, cfg.Address, "localhost")
	assert.Equal(t, cfg.AppEnv, "development")
	assert.Equal(t, cfg.LogLevel, int64(0))
	assert.Equal(t, cfg.StoreInterval, int64(10))
	assert.Equal(t, cfg.FileStoragePath, "/tmp")
	assert.Equal(t, cfg.Restore, false)
	assert.Equal(t, cfg.DatabaseDSN, "pg@pg")
}

func TestConfig_IsDBDSNPresent(t *testing.T) {
	type fields struct {
		DatabaseDSN string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "when present",
			want:   true,
			fields: fields{DatabaseDSN: "dsn"},
		},
		{
			name:   "when blank",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				DatabaseDSN: tt.fields.DatabaseDSN,
			}
			if got := c.IsDBDSNPresent(); got != tt.want {
				t.Errorf("Config.IsDBDSNPresent() = %v, want %v", got, tt.want)
			}
		})
	}
}
