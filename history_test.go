package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryAdd(t *testing.T) {
	h := &History{entries: make([]string, 0, maxHistoryItems)}

	h.Add("first")
	h.Add("second")
	h.Add("third")

	entries := h.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0] != "third" {
		t.Errorf("expected 'third' at [0], got %q", entries[0])
	}
	if entries[2] != "first" {
		t.Errorf("expected 'first' at [2], got %q", entries[2])
	}
}

func TestHistoryDedup(t *testing.T) {
	h := &History{entries: make([]string, 0, maxHistoryItems)}

	h.Add("hello")
	h.Add("world")
	h.Add("hello") // duplicate moves to front

	entries := h.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after dedup, got %d", len(entries))
	}
	if entries[0] != "hello" {
		t.Errorf("expected 'hello' at [0], got %q", entries[0])
	}
}

func TestHistoryMaxItems(t *testing.T) {
	h := &History{entries: make([]string, 0, maxHistoryItems)}

	for i := 0; i < maxHistoryItems+5; i++ {
		h.Add("entry " + string(rune('A'+i)))
	}

	entries := h.Entries()
	if len(entries) != maxHistoryItems {
		t.Fatalf("expected %d entries, got %d", maxHistoryItems, len(entries))
	}
}

func TestHistoryClear(t *testing.T) {
	h := &History{entries: make([]string, 0, maxHistoryItems)}

	h.Add("one")
	h.Add("two")
	h.Clear()

	entries := h.Entries()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestHistoryPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h1 := &History{entries: make([]string, 0, maxHistoryItems), path: path}
	h1.Add("persisted")
	h1.Add("data")

	h2 := &History{entries: make([]string, 0, maxHistoryItems), path: path}
	h2.load()

	entries := h2.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 persisted entries, got %d", len(entries))
	}
	if entries[0] != "data" {
		t.Errorf("expected 'data' at [0], got %q", entries[0])
	}
}

func TestHistoryPersistenceFileRemoved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h := &History{entries: make([]string, 0, maxHistoryItems), path: path}
	h.Add("test")
	os.Remove(path)

	h2 := &History{entries: make([]string, 0, maxHistoryItems), path: path}
	h2.load()
	if len(h2.Entries()) != 0 {
		t.Error("expected empty history after file removal")
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"hello world", 5, "hello world"},
		{"one two three four five six seven eight nine ten", 8, "one two three four five six seven eight…"},
		{"single", 3, "single"},
		{"", 5, ""},
		{"a b c", 3, "a b c"},
		{"a b c d", 3, "a b c…"},
	}

	for _, tt := range tests {
		got := TruncateText(tt.input, tt.max)
		if got != tt.expected {
			t.Errorf("TruncateText(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
		}
	}
}
