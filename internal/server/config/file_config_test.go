package config

import (
	"bytes"
	"io"
	"testing"
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
				source:          tt.fields.source,
				Address:         tt.fields.Address,
				StoreInterval:   tt.fields.StoreInterval,
				FileStoragePath: tt.fields.FileStoragePath,
				DatabaseDSN:     tt.fields.DatabaseDSN,
				PrivateKey:      tt.fields.PrivateKey,
			}
			if err := f.Configure(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("FileConfig.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
