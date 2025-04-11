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
)

// Config holds all configuration parameters for the server
type Config struct {
	Address         string    `json:"address" env:"ADDRESS"`                                                   // Server address and port
	AppEnv          string    `json:"app_env" env:"APP_ENV" envDefault:"development"`                          // Application environment
	LogLevel        int64     `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`                                // Logging level
	StoreInterval   int64     `json:"store_interval" env:"STORE_INTERVAL" envDefault:"300"`                    // Interval for storing metrics
	FileStoragePath string    `json:"file_storage_path" env:"FILE_STORAGE_PATH" envDefault:"data/records.txt"` // Path for file storage
	Restore         bool      `json:"restore" env:"RESTORE" envDefault:"true"`                                 // Whether to restore data on startup
	DatabaseDSN     string    `json:"database_dsn" env:"DATABASE_DSN"`                                         // Database connection string
	Key             string    `json:"key" env:"KEY"`                                                           // Secret key for request encryption
	PrivateKey      io.Reader `json:"private_key_path"`                                                        // Path to the private key file
}

// String returns a JSON string representation of the configuration
func (c Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// NewConfig creates a new Config instance by parsing command line flags and environment variables
func NewConfig() (Config, error) {
	c := Config{}
	if err := c.parseFlags(); err != nil {
		return Config{}, fmt.Errorf("internal/server/config.go parse flags error %w", err)
	}

	if err := c.prseEnvs(); err != nil {
		return Config{}, fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	return c, nil
}

const defaultStoreInterval = 300

// parseFlags parses command line flags into the Config struct
func (c *Config) parseFlags() error {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&c.StoreInterval, "i", defaultStoreInterval, "store data writer scheduler interval")
	flag.StringVar(&c.FileStoragePath, "f", "data/records.txt", "data storage path")
	flag.BoolVar(&c.Restore, "r", true, "flag - restore from file on boot")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn")
	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Secret key for http request encryption")
	}

	if flag.Lookup("crypto-key") == nil {
		var path string
		flag.StringVar(&path, "crypto-key", "", "path to the private key file")
		cert, err := prepareCert(path)
		if err != nil {
			return fmt.Errorf("internal/server/config.go parse flags error %w", err)
		}
		c.PrivateKey = cert
	}

	flag.Parse()

	return nil
}

// IsDBDSNPresent checks if a database connection string is configured
func (c *Config) IsDBDSNPresent() bool {
	return c.DatabaseDSN != ""
}

// prseEnvs parses environment variables into the Config struct
func (c *Config) prseEnvs() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	if val, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		if cert, err := prepareCert(val); err != nil {
			return fmt.Errorf("internal/server/config.go parse env error %w", err)
		} else {
			c.PrivateKey = cert
		}
	}

	return nil
}

func prepareCert(val string) (io.Reader, error) {
	if val == "" {
		return nil, nil
	}

	cert := &bytes.Buffer{}
	file, err := os.Open(val)
	if err != nil {
		return nil, fmt.Errorf("internal/server/config.go prepare cert error %w", err)
	}
	defer file.Close()

	_, err = io.Copy(cert, file)
	if err != nil {
		return nil, fmt.Errorf("internal/server/config.go prepare cert error %w", err)
	}
	return cert, nil
}
