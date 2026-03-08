package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Hotkey != "RightAlt" {
		t.Errorf("expected hotkey RightAlt, got %s", cfg.Hotkey)
	}
	if cfg.Backend != "local" {
		t.Errorf("expected backend local, got %s", cfg.Backend)
	}
	if cfg.DoubleTapMs != 300 {
		t.Errorf("expected double_tap_ms 300, got %d", cfg.DoubleTapMs)
	}
	if cfg.Local.ModelSize != "base" {
		t.Errorf("expected model_size base, got %s", cfg.Local.ModelSize)
	}
	if cfg.Local.Language != "auto" {
		t.Errorf("expected language auto, got %s", cfg.Local.Language)
	}
	if cfg.OpenAI.Language != "auto" {
		t.Errorf("expected openai language auto, got %s", cfg.OpenAI.Language)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Hotkey != "RightAlt" {
		t.Errorf("expected default hotkey, got %s", cfg.Hotkey)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "sussurai")
	os.MkdirAll(cfgDir, 0o755)

	content := []byte(`
hotkey = "ScrollLock"
backend = "openai"
double_tap_ms = 500

[local]
model_size = "tiny"
language = "pt"

[openai]
api_key = "sk-test123"
language = "en"
`)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), content, 0o644)

	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Hotkey != "ScrollLock" {
		t.Errorf("expected hotkey ScrollLock, got %s", cfg.Hotkey)
	}
	if cfg.Backend != "openai" {
		t.Errorf("expected backend openai, got %s", cfg.Backend)
	}
	if cfg.DoubleTapMs != 500 {
		t.Errorf("expected double_tap_ms 500, got %d", cfg.DoubleTapMs)
	}
	if cfg.Local.ModelSize != "tiny" {
		t.Errorf("expected model_size tiny, got %s", cfg.Local.ModelSize)
	}
	if cfg.Local.Language != "pt" {
		t.Errorf("expected language pt, got %s", cfg.Local.Language)
	}
	if cfg.OpenAI.APIKey != "sk-test123" {
		t.Errorf("expected api_key sk-test123, got %s", cfg.OpenAI.APIKey)
	}
}

func TestLoadConfigAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-from-env")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OpenAI.APIKey != "sk-from-env" {
		t.Errorf("expected api_key from env, got %s", cfg.OpenAI.APIKey)
	}
}

func TestModelPath(t *testing.T) {
	cfg := DefaultConfig()
	path, err := cfg.ModelPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
	if filepath.Base(path) != "ggml-base.bin" {
		t.Errorf("expected ggml-base.bin, got %s", filepath.Base(path))
	}
}

func TestModelPathExplicit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Local.ModelPath = "/custom/path/model.bin"
	path, err := cfg.ModelPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/custom/path/model.bin" {
		t.Errorf("expected custom path, got %s", path)
	}
}

func TestSaveConfigRoundtrip(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "sussurai")
	os.MkdirAll(cfgDir, 0o755)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := DefaultConfig()
	cfg.Backend = "groq"
	cfg.Groq.APIKey = "gsk-secret"
	cfg.Groq.Language = "pt"
	cfg.Groq.Translate = true
	cfg.OpenAI.APIKey = "sk-secret"

	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	// Read back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if loaded.Backend != "groq" {
		t.Errorf("expected backend groq, got %s", loaded.Backend)
	}
	if loaded.Groq.Language != "pt" {
		t.Errorf("expected groq language pt, got %s", loaded.Groq.Language)
	}
	if !loaded.Groq.Translate {
		t.Error("expected groq translate true")
	}
	// API keys should NOT be saved to config file.
	// Read the raw file to verify (env vars may fill them back on LoadConfig).
	raw, _ := os.ReadFile(filepath.Join(cfgDir, "config.toml"))
	if strings.Contains(string(raw), "gsk-secret") {
		t.Error("expected groq api_key to not be in saved config file")
	}
	if strings.Contains(string(raw), "sk-secret") {
		t.Error("expected openai api_key to not be in saved config file")
	}
}

func TestSaveConfigPreservesAllFields(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "sussurai")
	os.MkdirAll(cfgDir, 0o755)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := DefaultConfig()
	cfg.Hotkey = "ScrollLock"
	cfg.DoubleTapMs = 500
	cfg.Local.ModelSize = "small"
	cfg.Local.Language = "es"
	cfg.OpenAI.Language = "fr"
	cfg.OpenAI.Translate = true

	SaveConfig(cfg)
	loaded, _ := LoadConfig()

	if loaded.Hotkey != "ScrollLock" {
		t.Errorf("expected hotkey ScrollLock, got %s", loaded.Hotkey)
	}
	if loaded.DoubleTapMs != 500 {
		t.Errorf("expected double_tap_ms 500, got %d", loaded.DoubleTapMs)
	}
	if loaded.Local.ModelSize != "small" {
		t.Errorf("expected model_size small, got %s", loaded.Local.ModelSize)
	}
	if loaded.Local.Language != "es" {
		t.Errorf("expected local language es, got %s", loaded.Local.Language)
	}
	if loaded.OpenAI.Language != "fr" {
		t.Errorf("expected openai language fr, got %s", loaded.OpenAI.Language)
	}
	if !loaded.OpenAI.Translate {
		t.Error("expected openai translate true")
	}
}

func TestLoadVocabulary(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "sussurai")
	os.MkdirAll(cfgDir, 0o755)
	t.Setenv("XDG_CONFIG_HOME", dir)

	content := "# Comment\nJSON\nKubernetes\n\n# Another comment\nSussur.ai\n"
	os.WriteFile(filepath.Join(cfgDir, "vocabulary.txt"), []byte(content), 0o644)

	result := LoadVocabulary()
	if result != "JSON, Kubernetes, Sussur.ai" {
		t.Errorf("expected 'JSON, Kubernetes, Sussur.ai', got %q", result)
	}
}

func TestLoadVocabularyEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	result := LoadVocabulary()
	if result != "" {
		t.Errorf("expected empty vocabulary, got %q", result)
	}
}
