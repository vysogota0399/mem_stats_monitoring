package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type NewConfigOption func(*Config)

func NewConfig(pollInterval, reportInterval time.Duration, serverURL string) Config {
	c := Config{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		ServerURL:      serverURL,
	}

	if val, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if val, err := strconv.Atoi(val); err == nil {
			c.PollInterval = time.Duration(val) * time.Second
		}
	}

	if val, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if val, err := strconv.Atoi(val); err == nil {
			c.ReportInterval = time.Duration(val) * time.Second
		}
	}

	if val, ok := os.LookupEnv("ADDRESS"); ok {
		c.ServerURL = val
	}

	c.ServerURL = fmt.Sprintf("http://%s", c.ServerURL)
	return c
}

func (c *Config) String() string {
	conf := strings.Builder{}
	conf.WriteString(fmt.Sprintf("ServerURL: %s\n", c.ServerURL))
	conf.WriteString(fmt.Sprintf("PollInterval: %s\n", c.PollInterval))
	conf.WriteString(fmt.Sprintf("ReportInterval: %s\n", c.ReportInterval))
	return conf.String()
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
