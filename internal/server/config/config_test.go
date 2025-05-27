package config

import (
	"bytes"
	"io"
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
	err = os.Setenv("KEY", "")
	assert.NoError(t, err)
	err = os.Setenv("CRYPTO_KEY", "")
	assert.NoError(t, err)
	err = os.Setenv("CONFIG", "")
	assert.NoError(t, err)
	err = os.Setenv("TRUSTED_SUBNET", "127.0.0.1/8")
	assert.NoError(t, err)

	cfg, err := NewConfig(nil)
	assert.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Address)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, int64(0), cfg.LogLevel)
	assert.Equal(t, int64(10), cfg.StoreInterval)
	assert.Equal(t, "/tmp", cfg.FileStoragePath)
	assert.Equal(t, false, cfg.Restore)
	assert.Equal(t, "pg@pg", cfg.DatabaseDSN)
	assert.Equal(t, "", cfg.Key)
	assert.Equal(t, nil, cfg.PrivateKey)
	assert.Equal(t, "", cfg.ConfigPath)
	assert.Equal(t, "127.0.0.1/8", cfg.TrustedSubnet)

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

func Test_preparePrivateKey(t *testing.T) {
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
				val: "nonexistent.pem",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid private key file returns reader",
			args: args{
				val: "test.pem",
			},
			want:    bytes.NewReader([]byte("test private key")),
			wantErr: false,
			setup: func() error {
				return os.WriteFile("test.pem", []byte("test private key"), 0644)
			},
			cleanup: func() error {
				return os.Remove("test.pem")
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

			got, err := preparePrivateKey(tt.args.val)
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

func TestConfig_parseConfigFile(t *testing.T) {
	type fields struct {
		ConfigPath string
	}
	type args struct {
		fc FileConfigurer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func() error
		cleanup func() error
	}{
		{
			name: "empty config path",
			fields: fields{
				ConfigPath: "",
			},
			args: args{
				fc: nil,
			},
			wantErr: false,
		},
		{
			name: "non-existent config file",
			fields: fields{
				ConfigPath: "nonexistent.json",
			},
			args: args{
				fc: nil,
			},
			wantErr: true,
		},
		{
			name: "valid config file with file configurer",
			fields: fields{
				ConfigPath: "test_config.json",
			},
			args: args{
				fc: &FileConfig{},
			},
			wantErr: false,
			setup: func() error {
				return os.WriteFile("test_config.json", []byte(`{
					"address": "localhost:8080",
					"store_interval": 300,
					"file_storage_path": "/tmp/data",
					"database_dsn": "postgres://user:pass@localhost:5432/db"
				}`), 0644)
			},
			cleanup: func() error {
				return os.Remove("test_config.json")
			},
		},
		{
			name: "invalid config file",
			fields: fields{
				ConfigPath: "invalid_config.json",
			},
			args: args{
				fc: &FileConfig{},
			},
			wantErr: true,
			setup: func() error {
				return os.WriteFile("invalid_config.json", []byte("invalid json"), 0644)
			},
			cleanup: func() error {
				return os.Remove("invalid_config.json")
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

			c := &Config{
				ConfigPath: tt.fields.ConfigPath,
			}
			if err := c.parseConfigFile(tt.args.fc); (err != nil) != tt.wantErr {
				t.Errorf("Config.parseConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
