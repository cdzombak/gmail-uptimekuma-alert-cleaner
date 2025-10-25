package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"time"
)

// Config represents the application configuration
type Config struct {
	TimeWindow time.Duration `json:"-"`
}

// configFile represents the JSON structure of the config file
type configFile struct {
	TimeWindow string `json:"timeWindow"`
}

// LoadConfig loads the configuration from the config directory
// If the config file doesn't exist, returns default configuration
func LoadConfig(configDir string) (*Config, error) {
	configPath := path.Join(configDir, "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config
		return &Config{
			TimeWindow: 24 * time.Hour,
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}

	// Parse time window
	if cf.TimeWindow == "" {
		return nil, errors.New("timeWindow is required in config file")
	}

	timeWindow, err := time.ParseDuration(cf.TimeWindow)
	if err != nil {
		return nil, err
	}

	if timeWindow <= 0 {
		return nil, errors.New("timeWindow must be positive")
	}

	return &Config{
		TimeWindow: timeWindow,
	}, nil
}
