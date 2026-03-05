package main

import "fmt"

type GroqTranscriber = apiTranscriber

func NewGroqTranscriber(cfg GroqConfig) (*GroqTranscriber, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Groq API key not set. Set groq.api_key in config or GROQ_API_KEY env var")
	}
	return &apiTranscriber{
		apiKey:   cfg.APIKey,
		language: cfg.Language,
		endpoint: "https://api.groq.com/openai/v1/audio/transcriptions",
		model:    "whisper-large-v3",
	}, nil
}
