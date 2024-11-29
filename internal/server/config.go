package server

const (
	defaultAddress string = "localhost:8080"
)

type Config struct {
	address string
}

type NewConfigOption func(*Config)

func SetAddress(a string) NewConfigOption {
	return func(c *Config) {
		c.address = a
	}
}

func NewConfig(options ...NewConfigOption) Config {
	config := Config{
		address: defaultAddress,
	}
	for _, opt := range options {
		opt(&config)
	}

	return config
}
