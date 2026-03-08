package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bendahl/uinput"
)

// Virtual keyboard for simulating Ctrl+V after copying text to clipboard.
// We create it at startup and keep it alive so the compositor (COSMIC)
// has time to register it (~2s). This avoids the problem of ydotool type
// garbling UTF-8 characters.
type typerState struct {
	mu       sync.Mutex
	keyboard uinput.Keyboard
	ready    bool
}

var ts *typerState

func initTyper() error {
	kb, err := uinput.CreateKeyboard("/dev/uinput", []byte("sussurai-keyboard"))
	if err != nil {
		return fmt.Errorf("create uinput keyboard: %w", err)
	}

	ts = &typerState{keyboard: kb, ready: true}
	// Wait for compositor to register the virtual keyboard
	time.Sleep(2 * time.Second)
	return nil
}

func (t *typerState) PasteClipboard() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.ready {
		return fmt.Errorf("keyboard not ready")
	}

	// Simulate Ctrl+Shift+V (works in both terminals and GUI apps)
	if err := t.keyboard.KeyDown(uinput.KeyLeftctrl); err != nil {
		return fmt.Errorf("ctrl down: %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := t.keyboard.KeyDown(uinput.KeyLeftshift); err != nil {
		t.keyboard.KeyUp(uinput.KeyLeftctrl)
		return fmt.Errorf("shift down: %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := t.keyboard.KeyPress(uinput.KeyV); err != nil {
		t.keyboard.KeyUp(uinput.KeyLeftshift)
		t.keyboard.KeyUp(uinput.KeyLeftctrl)
		return fmt.Errorf("v press: %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := t.keyboard.KeyUp(uinput.KeyLeftshift); err != nil {
		t.keyboard.KeyUp(uinput.KeyLeftctrl)
		return fmt.Errorf("shift up: %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := t.keyboard.KeyUp(uinput.KeyLeftctrl); err != nil {
		return fmt.Errorf("ctrl up: %w", err)
	}

	return nil
}

func (t *typerState) Close() {
	if t.keyboard != nil {
		t.keyboard.Close()
	}
}

func PasteText(text string) error {
	// Step 1: Copy text to clipboard
	if err := copyToClipboard(text); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}

	// Step 2: Simulate Ctrl+V via pre-warmed uinput keyboard
	if ts != nil {
		if err := ts.PasteClipboard(); err != nil {
			fmt.Fprintf(os.Stderr, "[sussur.ai] uinput paste failed: %v\n", err)
		} else {
			return nil
		}
	}

	return fmt.Errorf("no paste method available")
}

func copyToClipboard(text string) error {
	sessionType := os.Getenv("XDG_SESSION_TYPE")

	var cmd *exec.Cmd
	if strings.Contains(sessionType, "wayland") {
		cmd = exec.Command("wl-copy")
	} else {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}

	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		// Try the other tool as fallback
		if strings.Contains(sessionType, "wayland") {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			cmd = exec.Command("wl-copy")
		}
		cmd.Stdin = strings.NewReader(text)
		if err2 := cmd.Run(); err2 != nil {
			return fmt.Errorf("%w / %w", err, err2)
		}
	}
	return nil
}

func CheckPasteDeps() error {
	// Check for clipboard tool
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if strings.Contains(sessionType, "wayland") {
		if _, err := exec.LookPath("wl-copy"); err != nil {
			return fmt.Errorf("wl-copy not found. Install: sudo apt install wl-clipboard")
		}
	} else {
		if _, err := exec.LookPath("xclip"); err != nil {
			return fmt.Errorf("xclip not found. Install: sudo apt install xclip")
		}
	}

	// Check uinput access
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("/dev/uinput not accessible. Run: sudo setfacl -m u:$USER:rw /dev/uinput")
	}
	f.Close()

	return nil
}
