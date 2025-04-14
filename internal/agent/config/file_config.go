package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
)

type FileConfig struct {
	source         io.Reader
	ServerURL      string `json:"address"`
	ReportInterval int64  `json:"report_interval"`
	PollInterval   int64  `json:"poll_interval"`
	HTTPCert       string `json:"crypto_key"`
}

func NewFromFile(path string) (*FileConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			zap.L().Error("config: failed to close file: %w", zap.Error(err))
		}
	}()

	return &FileConfig{
		source: file,
	}, nil
}

func NewFromReader(r io.Reader) (*FileConfig, error) {
	return &FileConfig{
		source: r,
	}, nil
}

func (f *FileConfig) Configure(c *Config) error {
	buffer := bufio.NewReader(f.source)

	if err := json.NewDecoder(buffer).Decode(f); err != nil {
		return fmt.Errorf("config: failed to decode file: %w", err)
	}

	if c.ServerURL == "" && f.ServerURL != "" {
		c.ServerURL = f.ServerURL
	}

	if c.ReportInterval == 0 && f.ReportInterval != 0 {
		c.ReportInterval = time.Duration(f.ReportInterval) * time.Second
	}

	if c.PollInterval == 0 && f.PollInterval != 0 {
		c.PollInterval = time.Duration(f.PollInterval) * time.Second
	}

	if c.HTTPCert == nil && f.HTTPCert != "" {
		targer, err := prepareCert(f.HTTPCert)
		if err != nil {
			return fmt.Errorf("config: failed to prepare cert: %w", err)
		}

		c.HTTPCert = targer
	}

	return nil
}
