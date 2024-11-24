package server

const (
	DefaultPort uint16 = 8080
	DefaultHost string = "localhost"
)

type Config struct {
	Host string
	Port uint16
}

type NewConfigOption func(*Config)

func SetHost(Host string) NewConfigOption {
	return func(c *Config) {
		c.Host = Host
	}
}

func SetPort(port uint16) NewConfigOption {
	return func(c *Config) {
		c.Port = port
	}
}

func NewConfig(options ...NewConfigOption) Config {
	config := Config{
		Host: DefaultHost,
		Port: DefaultPort,
	}
	for _, opt := range options {
		opt(&config)
	}

	return config
}
