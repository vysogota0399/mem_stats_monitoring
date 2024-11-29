package server

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Address string `env:"ADDRESS"`
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
