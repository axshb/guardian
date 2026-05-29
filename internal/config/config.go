package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port     string
	DBPath   string
	DataPath string
	LogLevel string
}

// Load reads configuration from the environment with sensible defaults.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/assets.db"
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "/app/data"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Port:     port,
		DBPath:   dbPath,
		DataPath: dataPath,
		LogLevel: logLevel,
	}, nil
}

// AppConfig represents the persisted JSON config in data/config.json.
type AppConfig struct {
	SetupComplete bool   `json:"setupComplete"`
	Theme         string `json:"theme"`
	AdminHash     string `json:"adminHash"`
}

// DefaultAppConfig returns sensible defaults.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		SetupComplete: false,
		Theme:         "gruvbox",
		AdminHash:     "",
	}
}

// LoadAppConfig reads the persisted config from data/config.json.
func LoadAppConfig(dataPath string) (*AppConfig, error) {
	path := filepath.Join(dataPath, "config.json")
	cfg := DefaultAppConfig()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Save defaults
			if saveErr := SaveAppConfig(dataPath, cfg); saveErr != nil {
				return cfg, fmt.Errorf("save default config: %w", saveErr)
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

// SaveAppConfig writes the config to data/config.json.
func SaveAppConfig(dataPath string, cfg *AppConfig) error {
	path := filepath.Join(dataPath, "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}
