package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestNewConfig(t *testing.T) {
	// Backup and restore environment variables
	envVars := []string{
		"OUTPUT",
		"LOG_LEVEL",
		"COUNTRY",
		"PROVINCE",
		"LOCALITY",
		"ORG",
		"ORG_UNIT",
		"TTL",
	}
	backup := make(map[string]string)
	for _, env := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			backup[env] = val
			if err := os.Unsetenv(env); err != nil {
				t.Fatalf("failed to unset env var %s: %v", env, err)
			}
		}
	}
	defer func() {
		for env, val := range backup {
			if err := os.Setenv(env, val); err != nil {
				t.Errorf("failed to restore env var %s: %v", env, err)
			}
		}
	}()

	// Backup and restore command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "default values",
			args: []string{"test"},
			want: &Config{
				LogLevel:   zapcore.Level(DefaultLogLevel),
				Ttl:        time.Hour * 24 * 365,
				OutputFile: "",
				Country:    "",
				Province:   "",
				Locality:   "",
				Org:        "",
				OrgUnit:    "",
				CommonName: "",
			},
			wantErr: false,
		},
		{
			name: "all flags set",
			args: []string{
				"test",
				"-o", "output.pem",
				"-ll", "-1", // DefaultLogLevel
				"-c", "US",
				"-p", "CA",
				"-l", "San Francisco",
				"-org", "Test Org",
				"-ou", "Test Unit",
				"-cn", "test.example.com",
				"-ttl", "24h",
			},
			want: &Config{
				LogLevel:   zapcore.Level(DefaultLogLevel),
				OutputFile: "output.pem",
				Country:    "US",
				Province:   "CA",
				Locality:   "San Francisco",
				Org:        "Test Org",
				OrgUnit:    "Test Unit",
				CommonName: "test.example.com",
				Ttl:        time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "all env vars set",
			args: []string{"test"},
			env: map[string]string{
				"OUTPUT":    "output.pem",
				"LOG_LEVEL": "0",
				"COUNTRY":   "US",
				"PROVINCE":  "CA",
				"LOCALITY":  "San Francisco",
				"ORG":       "Test Org",
				"ORG_UNIT":  "Test Unit",
				"TTL":       "24h",
			},
			want: &Config{
				LogLevel:   zapcore.InfoLevel,
				OutputFile: "output.pem",
				Country:    "US",
				Province:   "CA",
				Locality:   "San Francisco",
				Org:        "Test Org",
				OrgUnit:    "Test Unit",
				CommonName: "",
				Ttl:        time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			args: []string{"test"},
			env: map[string]string{
				"LOG_LEVEL": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid ttl",
			args: []string{"test"},
			env: map[string]string{
				"TTL": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for env, val := range tt.env {
				if err := os.Setenv(env, val); err != nil {
					t.Fatalf("failed to set env var %s: %v", env, err)
				}
			}

			// Set command line arguments
			os.Args = tt.args

			// Reset flag.CommandLine to avoid "flag redefined" errors
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			got, err := NewConfig()

			// Clean up environment variables
			for env := range tt.env {
				if unsererr := os.Unsetenv(env); unsererr != nil {
					t.Errorf("failed to unset env var %s: %v", env, unsererr)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
