package main

import (
	"strings"
	"testing"

	"github.com/grafov/evdev"
)

func TestResolveKeyCode(t *testing.T) {
	tests := []struct {
		name     string
		expected uint16
	}{
		{"RightAlt", evdev.KEY_RIGHTALT},
		{"LeftCtrl", evdev.KEY_LEFTCTRL},
		{"ScrollLock", evdev.KEY_SCROLLLOCK},
		{"Pause", evdev.KEY_PAUSE},
		{"CapsLock", evdev.KEY_CAPSLOCK},
		{"F13", evdev.KEY_F13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := resolveKeyCode(tt.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, code)
			}
		})
	}
}

func TestResolveKeyCodeUnknown(t *testing.T) {
	_, err := resolveKeyCode("NonExistentKey")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestSupportedKeys(t *testing.T) {
	keys := supportedKeys()
	if keys == "" {
		t.Error("expected non-empty supported keys string")
	}
	if !strings.Contains(keys, "RightAlt") {
		t.Error("expected supported keys to contain RightAlt")
	}
}
