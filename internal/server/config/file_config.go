package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

type FileConfig struct {
	source          io.Reader
	Address         string `json:"address"`
	StoreInterval   int64  `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	PrivateKey      string `json:"crypto_key"`
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

	if c.Address == "" && f.Address != "" {
		c.Address = f.Address
	}

	if c.StoreInterval == 0 && f.StoreInterval != 0 {
		c.StoreInterval = f.StoreInterval
	}

	if c.FileStoragePath == "" && f.FileStoragePath != "" {
		c.FileStoragePath = f.FileStoragePath
	}

	if c.DatabaseDSN == "" && f.DatabaseDSN != "" {
		c.DatabaseDSN = f.DatabaseDSN
	}

	if c.PrivateKey == nil && f.PrivateKey != "" {
		pk, err := preparePrivateKey(f.PrivateKey)
		if err != nil {
			return fmt.Errorf("config: failed to prepare private key: %w", err)
		}
		c.PrivateKey = pk
	}

	return nil
}
