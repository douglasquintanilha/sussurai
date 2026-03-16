package main

import (
	"fmt"
	"io"
	"strings"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type LocalTranscriber struct {
	model whisper.Model
	lang  string
}

func NewLocalTranscriber(cfg Config) (*LocalTranscriber, error) {
	modelPath, err := cfg.ModelPath()
	if err != nil {
		return nil, fmt.Errorf("model path: %w", err)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		if _, statErr := os.Stat(modelPath); os.IsNotExist(statErr) {
			return nil, fmt.Errorf("loading model: %s not found (did you run 'make download-model'?)", modelPath)
		}
		return nil, fmt.Errorf("loading model %s: %w", modelPath, err)
	}

	lang := cfg.Local.Language
	if lang == "" {
		lang = "auto"
	}

	return &LocalTranscriber{
		model: model,
		lang:  lang,
	}, nil
}

func (t *LocalTranscriber) Transcribe(samples []float32) (string, error) {
	ctx, err := t.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("creating context: %w", err)
	}

	if err := ctx.SetLanguage(t.lang); err != nil {
		return "", fmt.Errorf("setting language %q: %w", t.lang, err)
	}

	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return "", fmt.Errorf("processing audio: %w", err)
	}

	var parts []string
	for {
		segment, err := ctx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading segment: %w", err)
		}
		text := strings.TrimSpace(segment.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " "), nil
}

func (t *LocalTranscriber) Close() {
	t.model.Close()
}
