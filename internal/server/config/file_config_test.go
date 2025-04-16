package config

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig_Configure(t *testing.T) {
	type fields struct {
		source          io.Reader
		Address         string
		StoreInterval   int64
		FileStoragePath string
		DatabaseDSN     string
		PrivateKey      string
	}
	type args struct {
		c *Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid configuration",
			fields: fields{
				source: bytes.NewBufferString(`{
					"address": "localhost:8080",
					"store_interval": 300,
					"file_storage_path": "/tmp/data",
					"database_dsn": "postgres://user:pass@localhost:5432/db"
				}`),
			},
			args: args{
				c: &Config{},
			},
			wantErr: false,
		},
		{
			name: "invalid JSON",
			fields: fields{
				source: bytes.NewBufferString(`invalid json`),
			},
			args: args{
				c: &Config{},
			},
			wantErr: true,
		},
		{
			name: "partial configuration",
			fields: fields{
				source: bytes.NewBufferString(`{
					"address": "localhost:8080",
					"store_interval": 300
				}`),
			},
			args: args{
				c: &Config{},
			},
			wantErr: false,
		},
		{
			name: "empty configuration",
			fields: fields{
				source: bytes.NewBufferString(`{}`),
			},
			args: args{
				c: &Config{},
			},
			wantErr: false,
		},
		{
			name: "with existing config values",
			fields: fields{
				source: bytes.NewBufferString(`{
					"address": "localhost:8080",
					"store_interval": 300
				}`),
			},
			args: args{
				c: &Config{
					Address:       "existing:8080",
					StoreInterval: 200,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileConfig{
				Address:         tt.fields.Address,
				StoreInterval:   tt.fields.StoreInterval,
				FileStoragePath: tt.fields.FileStoragePath,
				DatabaseDSN:     tt.fields.DatabaseDSN,
				PrivateKey:      tt.fields.PrivateKey,
			}
			if err := f.Configure(tt.args.c, tt.fields.source); (err != nil) != tt.wantErr {
				t.Errorf("FileConfig.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFromFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *FileConfig
		wantErr bool
		setup   func() error
		cleanup func() error
	}{
		{
			name: "empty path",
			args: args{
				path: "",
			},
			want:    nil,	
			wantErr: true,
		},
		{
			name: "non-existent file",
			args: args{
				path: "nonexistent.json",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid config file",
			args: args{
				path: "test_config.json",
			},
			want:    &FileConfig{},
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
			args: args{
				path: "invalid_config.json",
			},
			want:    &FileConfig{},
			wantErr: false,
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

			got, err := NewFromFile(tt.args.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.IsType(t, tt.want, got)
		})
	}
}

func TestNewFromReader(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *FileConfig
		wantErr bool
	}{
		{
			name: "nil reader",
			args: args{
				r: nil,
			},
			want:    &FileConfig{},
			wantErr: false,
		},
		{
			name: "empty reader",
			args: args{
				r: bytes.NewBuffer([]byte{}),
			},
			want:    &FileConfig{},
			wantErr: false,
		},
		{
			name: "valid reader",
			args: args{
				r: bytes.NewBuffer([]byte(`{
					"address": "localhost:8080",
					"store_interval": 300,
					"file_storage_path": "/tmp/data",
					"database_dsn": "postgres://user:pass@localhost:5432/db"
				}`)),
			},
			want:    &FileConfig{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromReader(tt.args.r)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.IsType(t, tt.want, got)
		})
	}
}
