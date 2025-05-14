package config

import (
	"bytes"
	"io"
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

func Test_prepareCert(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    io.Reader
		wantErr bool
		setup   func() error
		cleanup func() error
	}{
		{
			name: "empty path returns nil reader",
			args: args{
				val: "",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "non-existent file returns error",
			args: args{
				val: "nonexistent.crt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid cert file returns reader",
			args: args{
				val: "test.crt",
			},
			want:    bytes.NewReader([]byte("test certificate")),
			wantErr: false,
			setup: func() error {
				return os.WriteFile("test.crt", []byte("test certificate"), 0644)
			},
			cleanup: func() error {
				return os.Remove("test.crt")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Errorf("cleanup failed: %v", err)
					}
				}()
			}

			got, err := prepareCert(tt.args.val)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			gotBytes, err := io.ReadAll(got)
			assert.NoError(t, err)

			wantBytes, err := io.ReadAll(tt.want)
			assert.NoError(t, err)

			assert.Equal(t, wantBytes, gotBytes)
		})
	}
}

func Test_fromFile(t *testing.T) {
	type args struct {
		cfg *Config
		fc  FileConfigurer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() error
		cleanup func() error
	}{
		{
			name: "empty config path",
			args: args{
				cfg: &Config{ConfigPath: ""},
				fc:  &FileConfig{},
			},
			wantErr: false,
		},
		{
			name: "non-existent config file",
			args: args{
				cfg: &Config{ConfigPath: "nonexistent.json"},
				fc:  &FileConfig{},
			},
			wantErr: true,
		},
		{
			name: "valid config file",
			args: args{
				cfg: &Config{ConfigPath: "test_config.json"},
				fc:  &FileConfig{},
			},
			wantErr: false,
			setup: func() error {
				return os.WriteFile("test_config.json", []byte(`{
					"address": "localhost:8080",
					"report_interval": 3,
					"poll_interval": 4
				}`), 0644)
			},
			cleanup: func() error {
				return os.Remove("test_config.json")
			},
		},
		{
			name: "invalid config file",
			args: args{
				cfg: &Config{ConfigPath: "invalid_config.json"},
				fc:  &FileConfig{},
			},
			wantErr: true,
			setup: func() error {
				return os.WriteFile("invalid_config.json", []byte("invalid json"), 0644)
			},
			cleanup: func() error {
				return os.Remove("invalid_config.json")
			},
		},
		{
			name: "file configurer error",
			args: args{
				cfg: &Config{ConfigPath: "test_config.json"},
				fc:  &FileConfig{},
			},
			wantErr: true,
			setup: func() error {
				return os.WriteFile("test_config.json", []byte(`{
					"address": "localhost:8080",
					"report_interval": "invalid",
					"poll_interval": 4
				}`), 0644)
			},
			cleanup: func() error {
				return os.Remove("test_config.json")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Errorf("cleanup failed: %v", err)
					}
				}()
			}

			err := fromFile(tt.args.cfg, tt.args.fc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
