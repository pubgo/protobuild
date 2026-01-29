// Package config provides file loading and saving utilities.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves configuration to a YAML file.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Vendor:   ".proto",
		Root:     []string{"proto"},
		Includes: []string{"proto", ".proto"},
	}
}
