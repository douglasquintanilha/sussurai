package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type apiTranscriber struct {
	apiKey    string
	language  string
	endpoint  string
	model     string
	translate bool
}

type OpenAITranscriber = apiTranscriber

func NewOpenAITranscriber(cfg OpenAIConfig) (*OpenAITranscriber, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set. Set openai.api_key in config or OPENAI_API_KEY env var")
	}
	endpoint := "https://api.openai.com/v1/audio/transcriptions"
	if cfg.Translate {
		endpoint = "https://api.openai.com/v1/audio/translations"
	}
	return &apiTranscriber{
		apiKey:    cfg.APIKey,
		language:  cfg.Language,
		endpoint:  endpoint,
		model:     "whisper-1",
		translate: cfg.Translate,
	}, nil
}

func (t *apiTranscriber) Transcribe(samples []float32) (string, error) {
	// Prepend ~300ms of silence to reduce hallucination on abrupt audio starts.
	// Whisper-large-v3 (Groq) is particularly sensitive to this.
	padSamples := whisperSampleRate * 3 / 10
	padded := make([]float32, padSamples+len(samples))
	copy(padded[padSamples:], samples)

	wavData := encodeWAV(padded, whisperSampleRate)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	if _, err := part.Write(wavData); err != nil {
		return "", fmt.Errorf("writing audio data: %w", err)
	}

	if err := writer.WriteField("model", t.model); err != nil {
		return "", fmt.Errorf("writing model field: %w", err)
	}

	if !t.translate && t.language != "" && t.language != "auto" {
		if err := writer.WriteField("language", t.language); err != nil {
			return "", fmt.Errorf("writing language field: %w", err)
		}
	}

	if err := writer.WriteField("temperature", "0"); err != nil {
		return "", fmt.Errorf("writing temperature field: %w", err)
	}

	if prompt := LoadVocabulary(); prompt != "" {
		if err := writer.WriteField("prompt", prompt); err != nil {
			return "", fmt.Errorf("writing prompt field: %w", err)
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", t.endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB max
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return result.Text, nil
}

func (t *apiTranscriber) Close() {}

// encodeWAV creates a WAV file in memory from float32 samples.
func encodeWAV(samples []float32, sampleRate int) []byte {
	numSamples := len(samples)
	dataSize := numSamples * 2 // 16-bit PCM
	fileSize := 36 + dataSize

	buf := make([]byte, 0, fileSize+8)
	put16 := func(v uint16) { buf = append(buf, byte(v), byte(v>>8)) }
	put32 := func(v uint32) { buf = append(buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24)) }

	// RIFF header
	buf = append(buf, "RIFF"...)
	put32(uint32(fileSize))
	buf = append(buf, "WAVE"...)

	// fmt chunk
	buf = append(buf, "fmt "...)
	put32(16)                      // chunk size
	put16(1)                       // PCM format
	put16(1)                       // mono
	put32(uint32(sampleRate))      // sample rate
	put32(uint32(sampleRate * 2))  // byte rate
	put16(2)                       // block align
	put16(16)                      // bits per sample

	// data chunk
	buf = append(buf, "data"...)
	put32(uint32(dataSize))

	for _, s := range samples {
		v := s
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		sample := int16(v * 32767)
		buf = append(buf, byte(sample), byte(sample>>8))
	}

	return buf
}
