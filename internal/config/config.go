package config

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Emojis         map[string]string            `yaml:"emojis"`
	Scopes         []string                     `yaml:"scopes"`
	BranchDefaults map[string]string            `yaml:"branch_defaults"` // branch pattern -> commit type
	RemoteConfig   string                       `yaml:"remote_config"`   // URL to remote config file
}

var DefaultConfig = Config{
	Emojis: map[string]string{
		"feat":     "✨",
		"fix":      "🐛",
		"docs":     "📚",
		"style":    "💎",
		"refactor": "📦",
		"perf":     "🚀",
		"test":     "🚨",
		"build":    "🛠",
		"ci":       "⚙️",
		"chore":    "♻️",
		"revert":   "🗑",
	},
}

// Load loads configuration from multiple sources and merges them.
// Priority: Repo config > User config > Remote config > Defaults
func Load(repoRoot string) (*Config, error) {
	cfg := DefaultConfig

	// Load User Config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, ".config", "git-sage", "config.yaml")
		_ = loadFromFile(userConfigPath, &cfg)
	}

	// Load Repo Config
	if repoRoot != "" {
		repoConfigPath := filepath.Join(repoRoot, ".git-sage.yaml")
		_ = loadFromFile(repoConfigPath, &cfg)
	}

	// Load Remote Config (lowest priority override after defaults, but before local)
	if cfg.RemoteConfig != "" {
		_ = loadFromURL(cfg.RemoteConfig, &cfg)
	}

	return &cfg, nil
}

func loadFromFile(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return decode(f, cfg)
}

func loadFromURL(url string, cfg *Config) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decode(resp.Body, cfg)
}

func decode(r io.Reader, cfg *Config) error {
	var fileCfg Config
	if err := yaml.NewDecoder(r).Decode(&fileCfg); err != nil {
		return err
	}
	for k, v := range fileCfg.Emojis {
		if cfg.Emojis == nil {
			cfg.Emojis = make(map[string]string)
		}
		cfg.Emojis[k] = v
	}
	if len(fileCfg.Scopes) > 0 {
		cfg.Scopes = append(cfg.Scopes, fileCfg.Scopes...)
	}
	if len(fileCfg.BranchDefaults) > 0 {
		if cfg.BranchDefaults == nil {
			cfg.BranchDefaults = make(map[string]string)
		}
		for k, v := range fileCfg.BranchDefaults {
			cfg.BranchDefaults[k] = v
		}
	}
	if fileCfg.RemoteConfig != "" {
		cfg.RemoteConfig = fileCfg.RemoteConfig
	}
	return nil
}
