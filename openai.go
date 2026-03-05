package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type apiTranscriber struct {
	apiKey   string
	language string
	endpoint string
	model    string
}

type OpenAITranscriber = apiTranscriber

func NewOpenAITranscriber(cfg OpenAIConfig) (*OpenAITranscriber, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set. Set openai.api_key in config or OPENAI_API_KEY env var")
	}
	return &apiTranscriber{
		apiKey:   cfg.APIKey,
		language: cfg.Language,
		endpoint: "https://api.openai.com/v1/audio/transcriptions",
		model:    "whisper-1",
	}, nil
}

func (t *apiTranscriber) Transcribe(samples []float32) (string, error) {
	wavData, err := encodeWAV(samples, whisperSampleRate)
	if err != nil {
		return "", fmt.Errorf("encoding WAV: %w", err)
	}

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

	if t.language != "" && t.language != "auto" {
		if err := writer.WriteField("language", t.language); err != nil {
			return "", fmt.Errorf("writing language field: %w", err)
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

	respBody, err := io.ReadAll(resp.Body)
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
func encodeWAV(samples []float32, sampleRate int) ([]byte, error) {
	numSamples := len(samples)
	dataSize := numSamples * 2 // 16-bit PCM
	fileSize := 36 + dataSize

	buf := bytes.NewBuffer(make([]byte, 0, fileSize+8))

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))            // chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))             // PCM format
	binary.Write(buf, binary.LittleEndian, uint16(1))             // mono
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))    // sample rate
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2))  // byte rate
	binary.Write(buf, binary.LittleEndian, uint16(2))             // block align
	binary.Write(buf, binary.LittleEndian, uint16(16))            // bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	for _, s := range samples {
		clamped := math.Max(-1, math.Min(1, float64(s)))
		binary.Write(buf, binary.LittleEndian, int16(clamped*32767))
	}

	return buf.Bytes(), nil
}
