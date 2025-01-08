package server

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	AppEnv   string `env:"APP_ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"Info"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Address: %s", c.Address)
}

func NewConfig(address string) Config {
	c := Config{
		Address: address,
	}

	if err := env.Parse(&c); err != nil {
		panic(err)
	}

	return c
}
