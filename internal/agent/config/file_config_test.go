package config

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig_Configure(t *testing.T) {
	type fields struct {
		ServerURL      string `json:"address"`
		ReportInterval int64  `json:"report_interval"`
		PollInterval   int64  `json:"poll_interval"`
		HTTPCert       string `json:"crypto_key"`
	}
	type args struct {
		c *Config
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T, target *Config, f *fields)
	}{
		{
			name: "don't replace empty fields",
			fields: fields{
				ServerURL:      "localhost:8080",
				ReportInterval: 3,
				PollInterval:   4,
			},
			args: args{
				c: &Config{
					ServerURL:      "127.0.0.1:8080",
					ReportInterval: 1,
					PollInterval:   2,
				},
			},
			assert: func(t *testing.T, target *Config, f *fields) {
				assert.NotEqual(t, f.ServerURL, target.ServerURL)
				assert.NotEqual(t, f.ReportInterval, target.ReportInterval)
				assert.NotEqual(t, f.PollInterval, target.PollInterval)
			},
		},
		{
			name: "replace empty fields",
			fields: fields{
				ServerURL:      "localhost:8080",
				ReportInterval: 3,
				PollInterval:   4,
			},
			args: args{
				c: &Config{},
			},
			assert: func(t *testing.T, target *Config, f *fields) {
				assert.Equal(t, f.ServerURL, target.ServerURL)
				assert.Equal(t, time.Duration(f.ReportInterval)*time.Second, target.ReportInterval)
				assert.Equal(t, time.Duration(f.PollInterval)*time.Second, target.PollInterval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(&tt.fields)
			assert.NoError(t, err)

			cfg := NewFileConfig()

			buff := bytes.NewBuffer(b)

			err = cfg.Configure(tt.args.c, buff)
			assert.NoError(t, err)

			tt.assert(t, tt.args.c, &tt.fields)
		})
	}
}
