package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/systray"
)

//go:embed icons/*.png
var iconFS embed.FS

var (
	trayReady  atomic.Bool
	trayQuitCh chan struct{}
)

var (
	iconIdle   []byte
	iconRec    []byte
	iconTransc []byte
	iconErr    []byte
)

// Backend menu
var (
	backendParent *systray.MenuItem
	backendItems  [3]*systray.MenuItem
	backendKeys   = [3]string{"groq", "openai", "local"}
	backendNames  = [3]string{"Groq", "OpenAI", "Local"}
)

// Language menu
var (
	langParent *systray.MenuItem
	langItems  [10]*systray.MenuItem
	langCodes  = [10]string{"auto", "en", "pt", "es", "fr", "de", "it", "ja", "zh", "ko"}
	langNames  = [10]string{"Auto Detect", "English", "Português", "Español", "Français", "Deutsch", "Italiano", "日本語", "中文", "한국어"}
)

// Translate toggle
var mTranslate *systray.MenuItem

// History menu
var (
	historyParent *systray.MenuItem
	historyItems  [maxHistoryItems]*systray.MenuItem
	historyTexts  [maxHistoryItems]string
	historyClear  *systray.MenuItem
	historyMu     sync.Mutex
)

func loadIcons() {
	iconIdle, _ = iconFS.ReadFile("icons/idle.png")
	iconRec, _ = iconFS.ReadFile("icons/recording.png")
	iconTransc, _ = iconFS.ReadFile("icons/transcribing.png")
	iconErr, _ = iconFS.ReadFile("icons/error.png")
}

func RunWithTray(appMain func(quit chan struct{})) {
	trayQuitCh = make(chan struct{})
	systray.Run(func() {
		onTrayReady()
		go appMain(trayQuitCh)
	}, func() {})
}

func onTrayReady() {
	loadIcons()
	systray.SetIcon(iconIdle)
	systray.SetTitle("Sussur.ai")
	systray.SetTooltip("Sussur.ai — Voice to Text")

	// Backend submenu
	backendParent = systray.AddMenuItem("Backend", "Choose transcription backend")
	for i, name := range backendNames {
		item := backendParent.AddSubMenuItem(name, "")
		backendItems[i] = item
		backend := backendKeys[i]
		go func() {
			for range item.ClickedCh {
				if err := SwitchBackend(backend); err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("Cannot switch: %v", err))
					hideAfter(3 * time.Second)
				} else {
					fmt.Printf("[sussur.ai] Switched to %s\n", backend)
				}
			}
		}()
	}

	// Language submenu
	langParent = systray.AddMenuItem("Language", "Choose input language")
	for i, name := range langNames {
		item := langParent.AddSubMenuItem(name, "")
		langItems[i] = item
		code := langCodes[i]
		go func() {
			for range item.ClickedCh {
				if err := SetLanguage(code); err != nil {
					SetOverlay(OverlayError, fmt.Sprintf("Language error: %v", err))
					hideAfter(3 * time.Second)
				} else {
					fmt.Printf("[sussur.ai] Language set to %s\n", code)
				}
			}
		}()
	}

	// Translate toggle
	mTranslate = systray.AddMenuItem("Translate to English", "Translate audio to English")
	go func() {
		for range mTranslate.ClickedCh {
			appCfgMu.RLock()
			current := appCfg.OpenAI.Translate || appCfg.Groq.Translate
			appCfgMu.RUnlock()
			if err := SetTranslate(!current); err != nil {
				SetOverlay(OverlayError, fmt.Sprintf("Translate error: %v", err))
				hideAfter(3 * time.Second)
			}
		}
	}()

	systray.AddSeparator()

	// History submenu
	historyParent = systray.AddMenuItem("History", "Recent transcriptions")
	historyParent.Disable()

	for i := 0; i < maxHistoryItems; i++ {
		item := historyParent.AddSubMenuItem("", "Click to paste")
		item.Hide()
		historyItems[i] = item

		idx := i
		go func() {
			for range item.ClickedCh {
				historyMu.Lock()
				text := historyTexts[idx]
				historyMu.Unlock()
				if text != "" {
					go func() {
						time.Sleep(200 * time.Millisecond)
						PasteText(text)
					}()
				}
			}
		}()
	}

	historyClear = historyParent.AddSubMenuItem("Clear History", "Clear all history")
	historyClear.Hide()
	go func() {
		for range historyClear.ClickedCh {
			if history != nil {
				history.Clear()
				RefreshHistoryMenu()
			}
		}
	}()

	systray.AddSeparator()

	mVocab := systray.AddMenuItem("Edit Vocabulary", "Edit vocabulary.txt to improve transcription accuracy")
	go func() {
		for range mVocab.ClickedCh {
			go openVocabularyFile()
		}
	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit Sussur.ai")
	go func() {
		<-mQuit.ClickedCh
		close(trayQuitCh)
		systray.Quit()
	}()

	trayReady.Store(true)

	if history != nil {
		RefreshHistoryMenu()
	}
}

// RefreshSettingsMenu updates the Backend, Language, and Translate checkmarks.
func RefreshSettingsMenu() {
	if !trayReady.Load() {
		return
	}

	appCfgMu.RLock()
	cfg := appCfg
	appCfgMu.RUnlock()

	// Backend — checkmark + visual prefix
	for i, key := range backendKeys {
		if key == cfg.Backend {
			backendItems[i].SetTitle("● " + backendNames[i])
			backendItems[i].Check()
		} else {
			backendItems[i].SetTitle("  " + backendNames[i])
			backendItems[i].Uncheck()
		}
	}

	// Language — checkmark + visual prefix
	var lang string
	switch cfg.Backend {
	case "openai":
		lang = cfg.OpenAI.Language
	case "groq":
		lang = cfg.Groq.Language
	default:
		lang = cfg.Local.Language
	}
	for i, code := range langCodes {
		if code == lang {
			langItems[i].SetTitle("● " + langNames[i])
			langItems[i].Check()
		} else {
			langItems[i].SetTitle("  " + langNames[i])
			langItems[i].Uncheck()
		}
	}

	// Translate — toggle title to show state
	var translate bool
	switch cfg.Backend {
	case "openai":
		translate = cfg.OpenAI.Translate
	case "groq":
		translate = cfg.Groq.Translate
	}
	if translate {
		mTranslate.SetTitle("Translate to English: ON")
		mTranslate.Check()
	} else {
		mTranslate.SetTitle("Translate to English: OFF")
		mTranslate.Uncheck()
	}
}

// RefreshHistoryMenu updates the tray history submenu from the current history entries.
func RefreshHistoryMenu() {
	if !trayReady.Load() {
		return
	}

	entries := history.Entries()

	historyMu.Lock()
	defer historyMu.Unlock()

	for i := 0; i < maxHistoryItems; i++ {
		if i < len(entries) {
			historyTexts[i] = entries[i]
			historyItems[i].SetTitle(TruncateText(entries[i], truncateWords))
			historyItems[i].Show()
		} else {
			historyItems[i].Hide()
			historyTexts[i] = ""
		}
	}

	if len(entries) > 0 {
		historyParent.Enable()
		historyClear.Show()
	} else {
		historyParent.Disable()
		historyClear.Hide()
	}
}

func openVocabularyFile() {
	dir, err := configDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[sussur.ai] config dir error: %v\n", err)
		return
	}

	path := filepath.Join(dir, "vocabulary.txt")

	// Create file with header if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700)
		os.WriteFile(path, []byte("# Vocabulary — one word or phrase per line\n# These help the transcriber recognize domain-specific terms\n# Example:\n# JSON\n# Kubernetes\n# Sussur.ai\n"), 0600)
	}

	cmd := exec.Command("xdg-open", path)
	cmd.Start()
}

func SetTrayState(state OverlayState) {
	if !trayReady.Load() {
		return
	}
	switch state {
	case OverlayRecording:
		systray.SetIcon(iconRec)
	case OverlayTranscribing:
		systray.SetIcon(iconTransc)
	case OverlayError:
		systray.SetIcon(iconErr)
	default:
		systray.SetIcon(iconIdle)
	}
}
