package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName    = "ccyolo"
	KeyringAccount = "anthropic-api-key"
)

type Config struct {
	Enabled  bool   `json:"enabled"`
	Preset   string `json:"preset"`
	Model    string `json:"model"`
	CacheTTL int    `json:"cache_ttl"`
	Logging  bool   `json:"logging"`
}

func DefaultConfig() Config {
	return Config{
		Enabled:  true,
		Preset:   "balanced",
		Model:    "claude-haiku-4-5-20251001",
		CacheTTL: 86400, // 24 hours
		Logging:  false,
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccyolo")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func CacheDir() string {
	return filepath.Join(ConfigDir(), "cache")
}

func Load() Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return cfg
	}

	json.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg Config) error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// API Key management

func GetAPIKey() string {
	// 1. Check env var first
	if key := os.Getenv("CCYOLO_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}

	// 2. Check keychain
	key, err := keyring.Get(ServiceName, KeyringAccount)
	if err == nil {
		return key
	}

	return ""
}

func SetAPIKey(key string) error {
	return keyring.Set(ServiceName, KeyringAccount, key)
}

func DeleteAPIKey() error {
	return keyring.Delete(ServiceName, KeyringAccount)
}

func HasAPIKey() bool {
	return GetAPIKey() != ""
}
