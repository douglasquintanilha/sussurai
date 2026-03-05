package main

// Transcriber is the interface for speech-to-text backends.
type Transcriber interface {
	// Transcribe converts audio samples (16kHz mono f32) to text.
	Transcribe(samples []float32) (string, error)
	// Close releases any resources held by the transcriber.
	Close()
}

// NewTranscriber creates a transcriber based on the config backend setting.
func NewTranscriber(cfg Config) (Transcriber, error) {
	switch cfg.Backend {
	case "openai":
		return NewOpenAITranscriber(cfg.OpenAI)
	case "groq":
		return NewGroqTranscriber(cfg.Groq)
	default:
		return NewLocalTranscriber(cfg)
	}
}
