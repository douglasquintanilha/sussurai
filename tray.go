package main

import (
	"embed"
	"sync/atomic"

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
	systray.SetTitle("Sussurai")
	systray.SetTooltip("Sussurai — Voice to Text")

	mQuit := systray.AddMenuItem("Quit", "Quit Sussurai")
	go func() {
		<-mQuit.ClickedCh
		close(trayQuitCh)
		systray.Quit()
	}()

	trayReady.Store(true)
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
