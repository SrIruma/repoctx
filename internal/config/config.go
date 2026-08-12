// Package config loads the optional repoctx.toml file that tunes scanning and
// context-file defaults across info, generate and audit.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// FileName is the default config file looked up in the target directory.
const FileName = "repoctx.toml"

// Config is the optional repoctx.toml file. Every field is optional; zero
// values mean "not set" and fall back to CLI flags or built-in defaults.
type Config struct {
	SkipDirs []string `toml:"skip_dirs"`
	MaxDepth int      `toml:"max_depth"`
	Files    []string `toml:"files"`
}

// Load reads FileName inside dir. A missing file returns a zero-value Config
// without error; a malformed file returns an error with the path in context.
func Load(dir string) (Config, error) {
	path := filepath.Join(dir, FileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return parse(path, data)
}

// LoadFile reads and parses a specific config file. Unlike Load, a missing
// file is an error: the caller asked for an explicit path.
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return parse(path, data)
}

func parse(path string, data []byte) (Config, error) {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}
