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

func main() {
	RunWithTray(appMain)
}

func appMain(quit chan struct{}) {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		systray.Quit()
		return
	}

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

	transcriber, err := NewTranscriber(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Transcriber init error: %v\n", err)
		systray.Quit()
		return
	}
	defer transcriber.Close()

	fmt.Println("[sussurai] Initializing virtual keyboard...")
	if err := initTyper(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: virtual keyboard failed: %v\n", err)
	}
	if ts != nil {
		defer ts.Close()
	}

	fmt.Printf("[sussurai] Ready! Hold %s to speak, release to paste.\n", cfg.Hotkey)
	fmt.Printf("[sussurai] Double-tap %s for toggle mode. Backend: %s\n", cfg.Hotkey, cfg.Backend)

	events := make(chan KeyEvent, 8)

	var closeOnce sync.Once
	shutdown := func() {
		closeOnce.Do(func() {
			fmt.Println("\n[sussurai] Shutting down...")
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
				text, err := transcriber.Transcribe(samples)
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

				SetOverlay(OverlaySuccess, text)
				hideAfter(1500 * time.Millisecond)
			}
		}
	}
}

func hideAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		SetOverlay(OverlayIdle, "")
	}()
}
