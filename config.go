package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

type Config struct {
	Hotkey      string       `toml:"hotkey"`
	Backend     string       `toml:"backend"`
	DoubleTapMs int          `toml:"double_tap_ms"`
	Local       LocalConfig  `toml:"local"`
	OpenAI      OpenAIConfig `toml:"openai"`
	Groq        GroqConfig   `toml:"groq"`
}

type LocalConfig struct {
	ModelPath string `toml:"model_path"`
	ModelSize string `toml:"model_size"`
	Language  string `toml:"language"`
}

type OpenAIConfig struct {
	APIKey    string `toml:"api_key"`
	Language  string `toml:"language"`
	Translate bool   `toml:"translate"`
}

type GroqConfig struct {
	APIKey    string `toml:"api_key"`
	Language  string `toml:"language"`
	Translate bool   `toml:"translate"`
}

func DefaultConfig() Config {
	return Config{
		Hotkey:      "RightAlt",
		Backend:     "local",
		DoubleTapMs: 300,
		Local: LocalConfig{
			ModelSize: "base",
			Language:  "auto",
		},
		OpenAI: OpenAIConfig{
			Language: "auto",
		},
		Groq: GroqConfig{
			Language: "auto",
		},
	}
}

func configDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, "sussurai"), nil
}

func dataDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "sussurai"), nil
}

func LoadConfig() (Config, error) {
	// Load .env from config dir or current dir (won't override existing env vars)
	if dir, err := configDir(); err == nil {
		godotenv.Load(filepath.Join(dir, ".env"))
	}
	godotenv.Load() // also try CWD

	cfg := DefaultConfig()

	dir, err := configDir()
	if err == nil {
		path := filepath.Join(dir, "config.toml")
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, &cfg); err != nil {
				return cfg, fmt.Errorf("parsing config: %w", err)
			}
		}
	}

	// Allow API keys from environment
	if cfg.OpenAI.APIKey == "" {
		cfg.OpenAI.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.Groq.APIKey == "" {
		cfg.Groq.APIKey = os.Getenv("GROQ_API_KEY")
	}

	return cfg, nil
}

func LoadVocabulary() string {
	dir, err := configDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, "vocabulary.txt"))
	if err != nil {
		return ""
	}
	// Join lines with ", " to form a prompt string
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var words []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}
	return strings.Join(words, ", ")
}

// SaveConfig writes the config to config.toml, omitting API keys.
func SaveConfig(cfg Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	// Never write API keys to config file — they belong in .env
	cfg.OpenAI.APIKey = ""
	cfg.Groq.APIKey = ""

	path := filepath.Join(dir, "config.toml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func (c *Config) ModelPath() (string, error) {
	if c.Local.ModelPath != "" {
		return c.Local.ModelPath, nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("ggml-%s.bin", c.Local.ModelSize)
	return filepath.Join(dir, "models", filename), nil
}
