package main

import (
	"encoding/json"
	"fmt"
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

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s\n\nPlease create a config file with the following format:\n{\n  \"root_dir\": \"/path/to/your/directory\"\n}", configPath)
	}
	
	// Try to read existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}
	
	return &config, nil
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