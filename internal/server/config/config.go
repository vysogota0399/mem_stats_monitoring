package config

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Address         string `json:"address" env:"ADDRESS"`
	AppEnv          string `json:"app_env" env:"APP_ENV" envDefault:"development"`
	LogLevel        string `json:"log_level" env:"LOG_LEVEL" envDefault:"Info"`
	StoreInterval   int64  `json:"store_interval" env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH" envDefault:"data/records.txt"`
	Restore         bool   `json:"restore" env:"RESTORE" envDefault:"true"`
}

func (c Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func NewConfig() (Config, error) {
	c := Config{}
	c.parseFlags()

	if err := c.prseEnvs(); err != nil {
		return Config{}, fmt.Errorf("internal/server/config.go parse env error %w", err)
	}

	return c, nil
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&c.StoreInterval, "i", 300, "store data writer scheduller interval")
	flag.StringVar(&c.FileStoragePath, "f", "data/records.txt", "data storage path")
	flag.BoolVar(&c.Restore, "r", true, "flat - restore from file on boot")
	flag.Parse()
}

func (c *Config) prseEnvs() error {
	return env.Parse(c)
}
