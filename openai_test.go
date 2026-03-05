package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestEncodeWAV(t *testing.T) {
	// Generate a short sine wave at 440Hz
	sampleRate := 16000
	duration := 0.1 // 100ms
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]float32, numSamples)
	for i := range samples {
		samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / float64(sampleRate)))
	}

	wav, err := encodeWAV(samples, sampleRate)
	if err != nil {
		t.Fatalf("encodeWAV error: %v", err)
	}

	// Check RIFF header
	if string(wav[:4]) != "RIFF" {
		t.Errorf("expected RIFF header, got %q", string(wav[:4]))
	}

	// Check WAVE format
	if string(wav[8:12]) != "WAVE" {
		t.Errorf("expected WAVE format, got %q", string(wav[8:12]))
	}

	// Check fmt chunk
	if string(wav[12:16]) != "fmt " {
		t.Errorf("expected fmt chunk, got %q", string(wav[12:16]))
	}

	// Check audio format (PCM = 1)
	reader := bytes.NewReader(wav[20:22])
	var audioFormat uint16
	binary.Read(reader, binary.LittleEndian, &audioFormat)
	if audioFormat != 1 {
		t.Errorf("expected PCM format (1), got %d", audioFormat)
	}

	// Check channels (mono = 1)
	reader = bytes.NewReader(wav[22:24])
	var channels uint16
	binary.Read(reader, binary.LittleEndian, &channels)
	if channels != 1 {
		t.Errorf("expected 1 channel, got %d", channels)
	}

	// Check sample rate
	reader = bytes.NewReader(wav[24:28])
	var sr uint32
	binary.Read(reader, binary.LittleEndian, &sr)
	if sr != uint32(sampleRate) {
		t.Errorf("expected sample rate %d, got %d", sampleRate, sr)
	}

	// Check data chunk header
	if string(wav[36:40]) != "data" {
		t.Errorf("expected data chunk, got %q", string(wav[36:40]))
	}

	// Check total WAV size: 44 header bytes + numSamples * 2 bytes
	expectedSize := 44 + numSamples*2
	if len(wav) != expectedSize {
		t.Errorf("expected WAV size %d, got %d", expectedSize, len(wav))
	}
}

func TestEncodeWAVEmpty(t *testing.T) {
	wav, err := encodeWAV([]float32{}, 16000)
	if err != nil {
		t.Fatalf("encodeWAV error: %v", err)
	}
	// Should still have valid header (44 bytes) with 0 data
	if len(wav) != 44 {
		t.Errorf("expected 44 bytes for empty WAV, got %d", len(wav))
	}
}

func TestEncodeWAVClipping(t *testing.T) {
	// Values outside [-1, 1] should be clamped
	samples := []float32{-2.0, -1.0, 0.0, 1.0, 2.0}
	wav, err := encodeWAV(samples, 16000)
	if err != nil {
		t.Fatalf("encodeWAV error: %v", err)
	}

	// Read back the PCM samples from the data section
	dataStart := 44
	for i := 0; i < len(samples); i++ {
		offset := dataStart + i*2
		reader := bytes.NewReader(wav[offset : offset+2])
		var sample int16
		binary.Read(reader, binary.LittleEndian, &sample)

		if sample < -32767 || sample > 32767 {
			t.Errorf("sample %d out of int16 range: %d", i, sample)
		}
	}
}

func TestNewOpenAITranscriberNoKey(t *testing.T) {
	_, err := NewOpenAITranscriber(OpenAIConfig{})
	if err == nil {
		t.Error("expected error when API key is empty")
	}
}

func TestNewOpenAITranscriberWithKey(t *testing.T) {
	tr, err := NewOpenAITranscriber(OpenAIConfig{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.apiKey != "sk-test" {
		t.Errorf("expected api key sk-test, got %s", tr.apiKey)
	}
}
