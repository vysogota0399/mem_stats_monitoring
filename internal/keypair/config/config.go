package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultLogLevel = -1
)

type Config struct {
	OutputFile string        `json:"output_file" env:"OUTPUT"`
	LogLevel   zapcore.Level `json:"log_level"`
	Country    string        `json:"country" env:"COUNTRY"`
	Province   string        `json:"province" env:"PROVINCE"`
	Locality   string        `json:"locality" env:"LOCALITY"`
	Org        string        `json:"org" env:"ORG"`
	OrgUnit    string        `json:"org_unit" env:"ORG_UNIT"`
	CommonName string        `json:"common_name"`
	Ttl        time.Duration `json:"ttl" env:"TTL"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		LogLevel: zapcore.Level(DefaultLogLevel),
	}

	var logLevel int

	if flag.Lookup("o") == nil {
		flag.StringVar(&cfg.OutputFile, "o", "", "output file")
	}

	if flag.Lookup("ll") == nil {
		flag.IntVar(&logLevel, "ll", DefaultLogLevel, "log level")
	}

	if flag.Lookup("c") == nil {
		flag.StringVar(&cfg.Country, "c", "", "country")
	}

	if flag.Lookup("p") == nil {
		flag.StringVar(&cfg.Province, "p", "", "province")
	}

	if flag.Lookup("l") == nil {
		flag.StringVar(&cfg.Locality, "l", "", "locality")
	}

	if flag.Lookup("org") == nil {
		flag.StringVar(&cfg.Org, "org", "", "organization")
	}

	if flag.Lookup("ou") == nil {
		flag.StringVar(&cfg.OrgUnit, "ou", "", "organization unit")
	}

	if flag.Lookup("cn") == nil {
		flag.StringVar(&cfg.CommonName, "cn", "", "common name")
	}

	if flag.Lookup("ttl") == nil {
		flag.DurationVar(&cfg.Ttl, "ttl", time.Hour*24*365, "certificate ttl")
	}

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: parse env error: %w", err)
	}

	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		ll, err := strconv.Atoi(logLevel)
		if err != nil {
			return nil, fmt.Errorf("config: parse log level error: %w", err)
		}
		cfg.LogLevel = zapcore.Level(ll)
	}

	return cfg, nil
}
