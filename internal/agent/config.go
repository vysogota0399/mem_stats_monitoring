package agent

import "time"

const (
	defaultReportInterval time.Duration = 2 * time.Second
	defaultPollInterval   time.Duration = 10 * time.Second
	defaultServerURL      string        = "http://localhost:8080"
)

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type NewConfigOption func(*Config)

func NewConfig(options ...NewConfigOption) Config {
	config := Config{
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		ServerURL:      defaultServerURL,
	}
	for _, opt := range options {
		opt(&config)
	}

	return config
}

func SetPollInterval(val time.Duration) NewConfigOption {
	return func(c *Config) {
		c.PollInterval = val
	}
}

func SetReportInterval(val time.Duration) NewConfigOption {
	return func(c *Config) {
		c.ReportInterval = val
	}
}

func SetServerURL(val string) NewConfigOption {
	return func(c *Config) {
		c.ServerURL = val
	}
}
