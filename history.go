package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	maxHistoryItems = 10
	truncateWords   = 8
)

var history *History

// History stores recent transcriptions in memory and persists to disk.
type History struct {
	mu      sync.Mutex
	entries []string
	path    string
}

func InitHistory() {
	h := &History{
		entries: make([]string, 0, maxHistoryItems),
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		h.path = filepath.Join(configDir, "sussurai", "history.json")
		h.load()
	}

	history = h
}

func (h *History) Add(text string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove duplicate if it already exists
	for i, e := range h.entries {
		if e == text {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}

	// Prepend new entry
	h.entries = append([]string{text}, h.entries...)

	// Trim to max size
	if len(h.entries) > maxHistoryItems {
		h.entries = h.entries[:maxHistoryItems]
	}

	h.save()
}

func (h *History) Entries() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]string, len(h.entries))
	copy(result, h.entries)
	return result
}

func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = h.entries[:0]
	h.save()
}

func (h *History) load() {
	data, err := os.ReadFile(h.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &h.entries)
	if len(h.entries) > maxHistoryItems {
		h.entries = h.entries[:maxHistoryItems]
	}
}

func (h *History) save() {
	if h.path == "" {
		return
	}
	data, err := json.Marshal(h.entries)
	if err != nil {
		return
	}
	os.WriteFile(h.path, data, 0600)
}

// TruncateText returns the first maxWords words followed by "…" if the text is longer.
func TruncateText(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}
	return strings.Join(words[:maxWords], " ") + "…"
}
