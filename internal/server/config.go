package server

const (
	defaultPort uint16 = 8080
	defaultHost string = "localhost"
)

type Config struct {
	host string
	port uint16
}

type NewConfigOption func(*Config)

func SetHost(host string) NewConfigOption {
	return func(c *Config) {
		c.host = host
	}
}

func SetPort(port uint16) NewConfigOption {
	return func(c *Config) {
		c.port = port
	}
}

func NewConfig(options ...NewConfigOption) Config {
	config := Config{
		host: defaultHost,
		port: defaultPort,
	}
	for _, opt := range options {
		opt(&config)
	}

	return config
}
