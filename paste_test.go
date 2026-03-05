package main

import (
	"testing"
)

func TestCheckPasteDepsDetectsSessionType(t *testing.T) {
	t.Setenv("XDG_SESSION_TYPE", "x11")
	_ = CheckPasteDeps()

	t.Setenv("XDG_SESSION_TYPE", "wayland")
	_ = CheckPasteDeps()
}

func TestCopyToClipboard(t *testing.T) {
	_ = copyToClipboard("test text")
}
