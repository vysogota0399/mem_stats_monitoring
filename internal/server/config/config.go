// Package config provides configuration management for the server component
package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FileConfigurer interface {
	Configure(c *Config, source io.Reader) error
}

// Config holds all configuration parameters for the server
type Config struct {
	Address         string    `json:"address" env:"ADDRESS"`                                                   // Server address and port
	AppEnv          string    `json:"app_env" env:"APP_ENV" envDefault:"development"`                          // Application environment
	LogLevel        int64     `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`                                // Logging level
	StoreInterval   int64     `json:"store_interval" env:"STORE_INTERVAL" envDefault:"15"`                     // Interval for storing metrics
	FileStoragePath string    `json:"file_storage_path" env:"FILE_STORAGE_PATH" envDefault:"data/records.txt"` // Path for file storage
	Restore         bool      `json:"restore" env:"RESTORE" envDefault:"true"`                                 // Whether to restore data on startup
	DatabaseDSN     string    `json:"database_dsn" env:"DATABASE_DSN"`                                         // Database connection string
	Key             string    `json:"key" env:"KEY"`                                                           // Secret key for request signature check
	PrivateKey      io.Reader `json:"private_key_path"`
	ConfigPath      string    `json:"config_path" env:"CONFIG" envDefault:""`
	TrustedSubnet   string    `json:"trusted_subnet" env:"TRUSTED_SUBNET" envDefault:""`
	GRPCPort        string    `json:"grpc_port" env:"GRPC_PORT" envDefault:"3200"`
}

func (c *Config) LLevel() zapcore.Level {
	return zapcore.Level(c.LogLevel)
}

// String returns a JSON string representation of the configuration
func (c Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// NewConfig creates a new Config instance by parsing command line flags and environment variables
func NewConfig(fc FileConfigurer) (*Config, error) {
	c := &Config{}
	if err := c.parseFlags(); err != nil {
		return nil, fmt.Errorf("internal/server/config.go parse flags error %w", err)
	}

	if err := c.parseEnvs(); err != nil {
		return nil, fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	if err := c.parseConfigFile(fc); err != nil {
		return nil, fmt.Errorf("internal/server/config.go parse config file error %w", err)
	}

	return c, nil
}

const defaultStoreInterval = 300

func (c *Config) parseConfigFile(fc FileConfigurer) error {
	if c.ConfigPath == "" {
		return nil
	}

	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("config: failed to open file %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			zap.L().Error("config: failed to close file", zap.Error(err))
		}
	}()

	if err := fc.Configure(c, file); err != nil {
		return fmt.Errorf("internal/server/config.go configure error %w", err)
	}

	return nil

}

// parseFlags parses command line flags into the Config struct
func (c *Config) parseFlags() error {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&c.StoreInterval, "i", defaultStoreInterval, "store data writer scheduler interval")
	flag.StringVar(&c.FileStoragePath, "f", "data/records.txt", "data storage path")
	flag.BoolVar(&c.Restore, "r", true, "flag - restore from file on boot")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn")
	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Secret key for http request signature check")
	}

	if flag.Lookup("crypto-key") == nil {
		var path string
		flag.StringVar(&path, "crypto-key", "", "path to the private key file")
		pk, err := preparePrivateKey(path)
		if err != nil {
			return fmt.Errorf("internal/server/config.go parse flags error %w", err)
		}
		c.PrivateKey = pk
	}

	if flag.Lookup("config") == nil {
		flag.StringVar(&c.ConfigPath, "config", "", "file to config.json")
	}

	if flag.Lookup("t") == nil {
		flag.StringVar(&c.TrustedSubnet, "t", "", "trusted subnet for incoming requests")
	}

	if flag.Lookup("grpc-port") == nil {
		flag.StringVar(&c.GRPCPort, "grpc-port", "3200", "grpc port")
	}

	flag.Parse()

	return nil
}

// IsDBDSNPresent checks if a database connection string is configured
func (c *Config) IsDBDSNPresent() bool {
	return c.DatabaseDSN != ""
}

// parseEnvs parses environment variables into the Config struct
func (c *Config) parseEnvs() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	if val, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		if pk, err := preparePrivateKey(val); err != nil {
			return fmt.Errorf("internal/server/config.go parse env error %w", err)
		} else {
			c.PrivateKey = pk
		}
	}

	return nil
}

func preparePrivateKey(val string) (io.Reader, error) {
	if val == "" {
		return nil, nil
	}

	cert := &bytes.Buffer{}
	file, err := os.Open(val)
	if err != nil {
		return nil, fmt.Errorf("internal/server/config.go prepare cert error %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			zap.L().Error("config: failed to close file: %w", zap.Error(closeErr))
		}
	}()

	_, err = io.Copy(cert, file)
	if err != nil {
		return nil, fmt.Errorf("internal/server/config.go prepare cert error %w", err)
	}
	return cert, nil
}
