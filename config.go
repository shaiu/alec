package main

import (
	"os"
	"path/filepath"
)

type Config struct {
	RootDir string
}

func getDefaultConfig() *Config {
	execPath, err := os.Executable()
	if err != nil {
		return &Config{RootDir: "."}
	}
	
	return &Config{
		RootDir: filepath.Dir(execPath),
	}
}

func (c *Config) GetRootDir() string {
	if c.RootDir == "" {
		return getDefaultConfig().RootDir
	}
	return c.RootDir
}