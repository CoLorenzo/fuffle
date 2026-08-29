package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configDir = "~/.config/fuffle"
const configFile = "~/.config/fuffle/config.json"

type ConfigFile struct {
	Port string `json:"port,omitempty"`
}

func configSet(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: config set requires key and value\n")
		fmt.Fprintf(os.Stderr, "Usage: %s config set <key> <value>\n", os.Args[0])
		os.Exit(1)
	}

	key := args[0]
	value := args[1]

	cfg, err := loadConfig()
	if err != nil {
		cfg = &ConfigFile{}
	}

	switch key {
	case "port":
		cfg.Port = value
	default:
		fmt.Fprintf(os.Stderr, "Unknown config key: %s\n", key)
		os.Exit(1)
	}

	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config set: %s = %s\n", key, value)
}

func loadConfig() (*ConfigFile, error) {
	expandedPath, err := expandPath(configFile)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, err
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func saveConfig(cfg *ConfigFile) error {
	expandedDir, err := expandPath(configDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(expandedDir, 0755); err != nil {
		return err
	}

	expandedPath, err := expandPath(configFile)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(expandedPath, data, 0644)
}

func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func loadPortFromConfig() string {
	cfg, err := loadConfig()
	if err != nil {
		return ""
	}
	return cfg.Port
}
