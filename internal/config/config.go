package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type DevConfig struct {
	Host    string `json:"host"`
	User    string `json:"user"`
	KeyPath string `json:"key_path"`
	WebRoot string `json:"web_root"`
}

func SaveDevConfig(cfg DevConfig) error {
	if err := os.MkdirAll(".now", 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(".now", "dev.json")
	return os.WriteFile(filePath, data, 0600)
}

func LoadDevConfig() (*DevConfig, error) {
	filePath := filepath.Join(".now", "dev.json")
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read .now/dev.json: %w", err)
	}

	var cfg DevConfig
	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		return nil, fmt.Errorf("invalid dev.json format: %w", err)
	}

	return &cfg, nil
}
