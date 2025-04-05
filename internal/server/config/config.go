// Package config provides configuration management for the server component
package config

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

// Config holds all configuration parameters for the server
type Config struct {
	Address         string `json:"address" env:"ADDRESS"`                                                   // Server address and port
	AppEnv          string `json:"app_env" env:"APP_ENV" envDefault:"development"`                          // Application environment
	LogLevel        int64  `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`                                // Logging level
	StoreInterval   int64  `json:"store_interval" env:"STORE_INTERVAL" envDefault:"300"`                    // Interval for storing metrics
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH" envDefault:"data/records.txt"` // Path for file storage
	Restore         bool   `json:"restore" env:"RESTORE" envDefault:"true"`                                 // Whether to restore data on startup
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`                                         // Database connection string
	Key             string `json:"key" env:"KEY"`                                                           // Secret key for request encryption
}

// String returns a JSON string representation of the configuration
func (c Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// NewConfig creates a new Config instance by parsing command line flags and environment variables
func NewConfig() (Config, error) {
	c := Config{}
	c.parseFlags()

	if err := c.prseEnvs(); err != nil {
		return Config{}, fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	return c, nil
}

const defaultStoreInterval = 300

// parseFlags parses command line flags into the Config struct
func (c *Config) parseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&c.StoreInterval, "i", defaultStoreInterval, "store data writer scheduler interval")
	flag.StringVar(&c.FileStoragePath, "f", "data/records.txt", "data storage path")
	flag.BoolVar(&c.Restore, "r", true, "flag - restore from file on boot")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn")
	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Secret key for http request encryption")
	}
	flag.Parse()
}

// IsDBDSNPresent checks if a database connection string is configured
func (c *Config) IsDBDSNPresent() bool {
	return c.DatabaseDSN != ""
}

// prseEnvs parses environment variables into the Config struct
func (c *Config) prseEnvs() error {
	return env.Parse(c)
}
