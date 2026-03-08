package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fyne.io/systray"
)

var (
	appCfg              Config
	appCfgMu            sync.RWMutex
	activeTranscriber   Transcriber
	activeTranscriberMu sync.RWMutex
	updateMu            sync.Mutex // serializes config updates to prevent TOCTOU races
)

func main() {
	InitHistory()
	RunWithTray(appMain)
}

func appMain(quit chan struct{}) {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		systray.Quit()
		return
	}

	appCfgMu.Lock()
	appCfg = cfg
	appCfgMu.Unlock()

	if err := CheckPasteDeps(); err != nil {
		fmt.Fprintf(os.Stderr, "Missing dependency: %v\n", err)
		systray.Quit()
		return
	}

	recorder, err := NewRecorder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Audio init error: %v\n", err)
		systray.Quit()
		return
	}
	defer recorder.Close()

	t, err := NewTranscriber(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Transcriber init error: %v\n", err)
		systray.Quit()
		return
	}

	activeTranscriberMu.Lock()
	activeTranscriber = t
	activeTranscriberMu.Unlock()
	defer func() {
		activeTranscriberMu.Lock()
		if activeTranscriber != nil {
			activeTranscriber.Close()
		}
		activeTranscriberMu.Unlock()
	}()

	RefreshSettingsMenu()

	fmt.Println("[sussur.ai] Initializing virtual keyboard...")
	if err := initTyper(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: virtual keyboard failed: %v\n", err)
	}
	if ts != nil {
		defer ts.Close()
	}

	fmt.Printf("[sussur.ai] Ready! Hold %s to speak, release to paste.\n", cfg.Hotkey)
	fmt.Printf("[sussur.ai] Double-tap %s for toggle mode. Backend: %s\n", cfg.Hotkey, cfg.Backend)

	events := make(chan KeyEvent, 8)

	var closeOnce sync.Once
	shutdown := func() {
		closeOnce.Do(func() {
			fmt.Println("\n[sussur.ai] Shutting down...")
			close(quit)
			systray.Quit()
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		shutdown()
	}()

	go func() {
		if err := ListenKeys(cfg.Hotkey, cfg.DoubleTapMs, events, quit); err != nil {
			fmt.Fprintf(os.Stderr, "Input error: %v\n", err)
			shutdown()
		}
	}()

	for {
		select {
		case <-quit:
			return
		case ev := <-events:
			switch ev.Type {
			case KeyCancel:
				recorder.Stop()
				SetOverlay(OverlayIdle, "")
				fmt.Println("[sussur.ai] Recording cancelled")

			case KeyPress:
				SetOverlay(OverlayRecording, "Recording...")
				if err := recorder.Start(); err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("Record error: %v", err))
				}

			case KeyRelease:
				samples, duration, err := recorder.Stop()
				if err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("%v", err))
					hideAfter(2 * time.Second)
					continue
				}

				if duration < 0.3 {
					SetOverlay(OverlayIdle, "")
					continue
				}

				SetOverlay(OverlayTranscribing, fmt.Sprintf("Transcribing %.1fs...", duration))

				activeTranscriberMu.RLock()
				text, err := activeTranscriber.Transcribe(samples)
				activeTranscriberMu.RUnlock()

				if err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("Error: %v", err))
					hideAfter(3 * time.Second)
					continue
				}

				if text == "" {
					SetOverlay(OverlayIdle, "")
					continue
				}

				SetOverlay(OverlayIdle, "")

				if err := PasteText(text); err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("Paste error: %v", err))
					hideAfter(3 * time.Second)
					continue
				}

				history.Add(text)
				RefreshHistoryMenu()

				SetOverlay(OverlaySuccess, text)
				hideAfter(1500 * time.Millisecond)
			}
		}
	}
}

// updateConfig atomically updates the config, rebuilds the transcriber, and saves.
func updateConfig(modify func(*Config)) error {
	updateMu.Lock()
	defer updateMu.Unlock()

	appCfgMu.RLock()
	newCfg := appCfg
	appCfgMu.RUnlock()

	modify(&newCfg)

	newT, err := NewTranscriber(newCfg)
	if err != nil {
		return err
	}

	appCfgMu.Lock()
	appCfg = newCfg
	appCfgMu.Unlock()

	activeTranscriberMu.Lock()
	old := activeTranscriber
	activeTranscriber = newT
	activeTranscriberMu.Unlock()

	if old != nil {
		old.Close()
	}

	if err := SaveConfig(newCfg); err != nil {
		fmt.Fprintf(os.Stderr, "[sussur.ai] Warning: could not save config: %v\n", err)
	}
	RefreshSettingsMenu()
	return nil
}

// SwitchBackend changes the active transcription backend at runtime.
func SwitchBackend(backend string) error {
	return updateConfig(func(cfg *Config) {
		cfg.Backend = backend
	})
}

// SetLanguage changes the input language for all backends.
func SetLanguage(lang string) error {
	return updateConfig(func(cfg *Config) {
		cfg.Local.Language = lang
		cfg.OpenAI.Language = lang
		cfg.Groq.Language = lang
	})
}

// SetTranslate toggles the translate-to-English mode.
func SetTranslate(enabled bool) error {
	return updateConfig(func(cfg *Config) {
		cfg.OpenAI.Translate = enabled
		cfg.Groq.Translate = enabled
	})
}

func hideAfter(d time.Duration) {
	gen := overlayGen.Load()
	go func() {
		time.Sleep(d)
		// Only reset to idle if no new state was set since we were scheduled
		if overlayGen.Load() == gen {
			SetOverlay(OverlayIdle, "")
		}
	}()
}
