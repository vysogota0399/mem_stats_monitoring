package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FileConfigurer interface {
	Configure(c *Config, source io.Reader) error
}

type Config struct {
	ServerURL      string        `json:"server_url"`
	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	LogLevel       int64         `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`
	Key            string        `json:"key" env:"KEY"`
	RateLimit      int           `json:"rate_limit" env:"RATE_LIMIT"`
	ProfileAddress string        `json:"profile_address" env:"PROFILE_ADDRESS"`
	MaxAttempts    uint8         `json:"max_attempts" env:"MAX_ATTEMPTS" envDefault:"5"`
	HTTPCert       io.Reader     `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigPath     string        `json:"config_path" env:"CONFIG" envDefault:""`
}

func NewConfig(f FileConfigurer) (Config, error) {
	c := Config{}
	if err := c.parseFlags(); err != nil {
		return Config{}, fmt.Errorf("config: failed to parse flags: %w", err)
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

	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		llvl, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return c, err
		}
		c.LogLevel = llvl
	} else {
		c.LogLevel = 0
	}

	if key, ok := os.LookupEnv("KEY"); ok {
		c.Key = key
	}

	if configPaht, ok := os.LookupEnv("CONFIG"); ok {
		c.ConfigPath = configPaht
	}

	if val, ok := os.LookupEnv("RATE_LIMIT"); ok {
		rLimit, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return c, err
		}
		c.RateLimit = int(rLimit)
	}

	if val, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		if cert, err := prepareCert(val); err != nil {
			return c, fmt.Errorf("config: failed to prepare cert: %w", err)
		} else {
			c.HTTPCert = cert
		}
	}

	if val, ok := os.LookupEnv("MAX_ATTEMPTS"); ok {
		maxAttempts, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return c, err
		}
		c.MaxAttempts = uint8(maxAttempts)
	} else {
		c.MaxAttempts = 5
	}

	c.ServerURL = fmt.Sprintf("http://%s", c.ServerURL)

	if err := fromFile(&c, f); err != nil {
		return Config{}, err
	}

	return c, nil
}

func (c *Config) LLevel() zapcore.Level {
	return zapcore.Level(c.LogLevel)
}

func (c *Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Config) parseFlags() error {
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

	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Secret key form http request encryption")
	}

	if flag.Lookup("l") == nil {
		flag.IntVar(&c.RateLimit, "l", runtime.GOMAXPROCS(0), "Reporter worker pool limit")
	}

	if flag.Lookup("crypto-key") == nil {
		var certPath string
		flag.StringVar(&certPath, "crypto-key", "", "Crypto key for encryption")
		cert, err := prepareCert(certPath)
		if err != nil {
			return fmt.Errorf("config: failed to prepare cert: %w", err)
		}
		c.HTTPCert = cert
	}

	if flag.Lookup("c") == nil {
		flag.StringVar(&c.ConfigPath, "c", "", "file to config.json")
	}

	flag.Parse()

	c.PollInterval = time.Duration(pollInterval) * time.Second
	c.ReportInterval = time.Duration(reportInterval) * time.Second

	return nil
}

func prepareCert(val string) (io.Reader, error) {
	if val == "" {
		return nil, nil
	}

	cert := &bytes.Buffer{}
	file, err := os.Open(val)
	if err != nil {
		return nil, fmt.Errorf("config: failed to open cert: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			zap.L().Error("config: failed to close cert: %w", zap.Error(closeErr))
		}
	}()

	_, err = io.Copy(cert, file)
	if err != nil {
		return nil, fmt.Errorf("config: failed to copy cert: %w", err)
	}

	return cert, nil
}

func fromFile(cfg *Config, fc FileConfigurer) error {
	if cfg.ConfigPath == "" {
		return nil
	}

	file, err := os.Open(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("config: failed to open file %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			zap.L().Error("config: failed to close file", zap.Error(err))
		}
	}()

	if err := fc.Configure(cfg, file); err != nil {
		return fmt.Errorf("config: failed to configure from file: %w", err)
	}

	return nil
}
