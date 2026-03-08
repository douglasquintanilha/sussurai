package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grafov/evdev"
)

type KeyEventType int

const (
	KeyPress   KeyEventType = iota
	KeyRelease
	KeyCancel
)

type KeyEvent struct {
	Type KeyEventType
}

var keyNameToCode = map[string]uint16{
	"RightAlt":   evdev.KEY_RIGHTALT,
	"LeftAlt":    evdev.KEY_LEFTALT,
	"RightCtrl":  evdev.KEY_RIGHTCTRL,
	"LeftCtrl":   evdev.KEY_LEFTCTRL,
	"ScrollLock": evdev.KEY_SCROLLLOCK,
	"Pause":      evdev.KEY_PAUSE,
	"CapsLock":   evdev.KEY_CAPSLOCK,
	"RightShift": evdev.KEY_RIGHTSHIFT,
	"LeftShift":  evdev.KEY_LEFTSHIFT,
	"RightMeta":  evdev.KEY_RIGHTMETA,
	"LeftMeta":   evdev.KEY_LEFTMETA,
	"F13":        evdev.KEY_F13,
	"F14":        evdev.KEY_F14,
	"F15":        evdev.KEY_F15,
}

func resolveKeyCode(name string) (uint16, error) {
	code, ok := keyNameToCode[name]
	if !ok {
		return 0, fmt.Errorf("unknown key name: %q (supported: %s)", name, supportedKeys())
	}
	return code, nil
}

func supportedKeys() string {
	keys := make([]string, 0, len(keyNameToCode))
	for k := range keyNameToCode {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func findKeyboards() ([]*evdev.InputDevice, error) {
	matches, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		return nil, err
	}

	var keyboards []*evdev.InputDevice
	for _, path := range matches {
		dev, err := evdev.Open(path)
		if err != nil {
			continue
		}
		if _, hasKey := dev.Capabilities["EV_KEY"]; hasKey {
			keyboards = append(keyboards, dev)
		} else {
			dev.File.Close()
		}
	}

	if len(keyboards) == 0 {
		return nil, fmt.Errorf("no keyboard devices found. Are you in the 'input' group? Try: sudo usermod -aG input %s", os.Getenv("USER"))
	}
	return keyboards, nil
}

// ListenKeys monitors all keyboards for the configured hotkey and sends events.
func ListenKeys(hotkey string, doubleTapMs int, events chan<- KeyEvent, quit <-chan struct{}) error {
	keyCode, err := resolveKeyCode(hotkey)
	if err != nil {
		return err
	}

	keyboards, err := findKeyboards()
	if err != nil {
		return err
	}
	defer func() {
		for _, kb := range keyboards {
			kb.File.Close()
		}
	}()

	fmt.Printf("[sussur.ai] Monitoring %d keyboard device(s) for %s\n", len(keyboards), hotkey)

	rawEvents := make(chan evdev.InputEvent, 64)
	for _, kb := range keyboards {
		go func(dev *evdev.InputDevice) {
			for {
				evts, err := dev.Read()
				if err != nil {
					return
				}
				for _, ev := range evts {
					if ev.Type == uint16(evdev.EV_KEY) && (ev.Code == keyCode || ev.Code == evdev.KEY_ESC) {
						select {
						case rawEvents <- ev:
						case <-quit:
							return
						}
					}
				}
			}
		}(kb)
	}

	doubleTapDuration := time.Duration(doubleTapMs) * time.Millisecond
	var lastReleaseTime time.Time
	toggleMode := false
	recording := false

	for {
		select {
		case <-quit:
			return nil
		case ev := <-rawEvents:
			// ESC cancels recording
			if ev.Code == evdev.KEY_ESC && ev.Value == 1 && recording {
				recording = false
				toggleMode = false
				events <- KeyEvent{Type: KeyCancel}
				continue
			}

			switch ev.Value {
			case 1: // Key down
				if !recording {
					timeSinceRelease := time.Since(lastReleaseTime)
					if timeSinceRelease < doubleTapDuration && !lastReleaseTime.IsZero() {
						toggleMode = true
						fmt.Println("[sussur.ai] Toggle mode: recording (tap again to stop)")
					} else {
						toggleMode = false
					}
					recording = true
					events <- KeyEvent{Type: KeyPress}
				} else if toggleMode {
					recording = false
					toggleMode = false
					events <- KeyEvent{Type: KeyRelease}
				}
			case 0: // Key up
				if recording && !toggleMode {
					recording = false
					events <- KeyEvent{Type: KeyRelease}
				}
				lastReleaseTime = time.Now()
			}
		}
	}
}
