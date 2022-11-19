package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bananenpro/cli"
	"github.com/adrg/xdg"
)

type Config struct {
	// The codegame-share instance to use.
	ShareURL string `json:"share_url"`
	// The port to use for `codegame run`.
	DevPort int `json:"dev_port"`
}

type config struct {
	ShareURL *string `json:"share_url"`
	DevPort  *int    `json:"dev_port"`
}

var Default = Config{
	ShareURL: "share.code-game.org",
	DevPort:  8080,
}

var configDir = filepath.Join(xdg.ConfigHome, "codegame")

// Load reads the config from disk and returns a Config object with all unset fields set to their default.
func Load() Config {
	file, err := os.Open(filepath.Join(configDir, "config.json"))
	if err != nil {
		return Default
	}
	defer file.Close()

	var c config
	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		cli.Warn("Failed to decode config file: %s", err)
		return Default
	}

	ret := Default
	if c.ShareURL != nil {
		ret.ShareURL = *c.ShareURL
	}
	if c.DevPort != nil {
		ret.DevPort = *c.DevPort
	}
	return ret
}

// Save creates or overwrites the config file with the values of self.
func (c Config) Save() error {
	err := os.MkdirAll(configDir, 0o775)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(filepath.Join(configDir, "config.json"))
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(c)
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}
	return nil
}
