package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerURL      string        `json:"server_url"`
	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	LogLevel       int64         `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`
}

func NewConfig() (Config, error) {
	c := Config{}
	c.parseFlags()

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

	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		llvl, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return c, err
		}
		c.LogLevel = llvl
	} else {
		c.LogLevel = 0
	}

	c.ServerURL = fmt.Sprintf("http://%s", c.ServerURL)
	return c, nil
}

func (c *Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Config) parseFlags() {
	var (
		pollInterval   int64
		reportInterval int64
	)

	const (
		defaultReportIntercal = 10
		defaultPollInterval   = 2
	)

	if flag.Lookup("a") == nil {
		flag.StringVar(&c.ServerURL, "a", "localhost:8080", "address and port to run server")
	}

	if flag.Lookup("p") == nil {
		flag.Int64Var(&pollInterval, "p", defaultPollInterval, "Poll interval")
	}

	if flag.Lookup("r") == nil {
		flag.Int64Var(&reportInterval, "r", defaultReportIntercal, "Report interval")
	}

	flag.Parse()

	c.PollInterval = time.Duration(pollInterval) * time.Second
	c.ReportInterval = time.Duration(reportInterval) * time.Second
}
