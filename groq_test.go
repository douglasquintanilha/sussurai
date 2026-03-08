package main

import (
	"strings"
	"testing"
)

func TestNewGroqTranscriberNoKey(t *testing.T) {
	_, err := NewGroqTranscriber(GroqConfig{})
	if err == nil {
		t.Error("expected error when API key is empty")
	}
}

func TestNewGroqTranscriberWithKey(t *testing.T) {
	tr, err := NewGroqTranscriber(GroqConfig{APIKey: "gsk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.apiKey != "gsk-test" {
		t.Errorf("expected api key gsk-test, got %s", tr.apiKey)
	}
	if tr.model != "whisper-large-v3" {
		t.Errorf("expected model whisper-large-v3, got %s", tr.model)
	}
}

func TestGroqTranscriptionEndpoint(t *testing.T) {
	tr, _ := NewGroqTranscriber(GroqConfig{APIKey: "gsk-test"})
	if !strings.Contains(tr.endpoint, "/transcriptions") {
		t.Errorf("expected transcriptions endpoint, got %s", tr.endpoint)
	}
}

func TestGroqTranslationEndpoint(t *testing.T) {
	tr, _ := NewGroqTranscriber(GroqConfig{APIKey: "gsk-test", Translate: true})
	if !strings.Contains(tr.endpoint, "/translations") {
		t.Errorf("expected translations endpoint, got %s", tr.endpoint)
	}
}

func TestOpenAITranscriptionEndpoint(t *testing.T) {
	tr, _ := NewOpenAITranscriber(OpenAIConfig{APIKey: "sk-test"})
	if !strings.Contains(tr.endpoint, "/transcriptions") {
		t.Errorf("expected transcriptions endpoint, got %s", tr.endpoint)
	}
}

func TestOpenAITranslationEndpoint(t *testing.T) {
	tr, _ := NewOpenAITranscriber(OpenAIConfig{APIKey: "sk-test", Translate: true})
	if !strings.Contains(tr.endpoint, "/translations") {
		t.Errorf("expected translations endpoint, got %s", tr.endpoint)
	}
}

func TestNewTranscriberFactoryNoKeys(t *testing.T) {
	// API backends should fail without keys
	for _, backend := range []string{"groq", "openai"} {
		t.Run(backend, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Backend = backend
			_, err := NewTranscriber(cfg)
			if err == nil {
				t.Errorf("expected error for %s without API key", backend)
			}
		})
	}
}

func TestNewTranscriberFactoryWithKeys(t *testing.T) {
	for _, tt := range []struct {
		backend string
		setup   func(*Config)
	}{
		{"groq", func(c *Config) { c.Groq.APIKey = "gsk-test" }},
		{"openai", func(c *Config) { c.OpenAI.APIKey = "sk-test" }},
	} {
		t.Run(tt.backend, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Backend = tt.backend
			tt.setup(&cfg)
			tr, err := NewTranscriber(cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tr.Close()
		})
	}
}
