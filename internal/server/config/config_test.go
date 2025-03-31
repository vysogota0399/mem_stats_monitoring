package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	os.Setenv("ADDRESS", "localhost")
	os.Setenv("APP_ENV", "development")
	os.Setenv("LOG_LEVEL", "0")
	os.Setenv("STORE_INTERVAL", "10")
	os.Setenv("FILE_STORAGE_PATH", "/tmp")
	os.Setenv("RESTORE", "false")
	os.Setenv("DATABASE_DSN", "pg@pg")

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
