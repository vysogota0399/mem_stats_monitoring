package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type FileConfig struct {
	Address         string `json:"address"`
	StoreInterval   int64  `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	PrivateKey      string `json:"crypto_key"`
}

func NewFileConfig() *FileConfig {
	return &FileConfig{}
}

func (f *FileConfig) Configure(c *Config, source io.Reader) error {
	buffer := bufio.NewReader(source)

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
