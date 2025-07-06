package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	RootDir string `json:"root_dir"`
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".alec.json"
	}
	return filepath.Join(homeDir, ".alec.json")
}

func loadConfig() *Config {
	configPath := getConfigPath()
	
	// Try to read existing config file
	if data, err := os.ReadFile(configPath); err == nil {
		var config Config
		if json.Unmarshal(data, &config) == nil {
			return &config
		}
	}
	
	// Return default config if file doesn't exist or can't be read
	return getDefaultConfig()
}

func (c *Config) Save() error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func getDefaultConfig() *Config {
	return &Config{
		RootDir: ".",
	}
}

func (c *Config) GetRootDir() string {
	if c.RootDir == "" {
		return getDefaultConfig().RootDir
	}
	return c.RootDir
}