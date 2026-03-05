package main

import (
	"os"
	"path/filepath"
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
