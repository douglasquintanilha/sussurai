package main

import "fmt"

type OverlayState int

const (
	OverlayIdle         OverlayState = 0
	OverlayRecording    OverlayState = 1
	OverlayTranscribing OverlayState = 2
	OverlaySuccess      OverlayState = 3
	OverlayError        OverlayState = 4
)

func SetOverlay(state OverlayState, text string) {
	SetTrayState(state)
	if text == "" {
		return
	}
	prefix := "[sussurai]"
	switch state {
	case OverlayRecording:
		prefix = "[sussurai] \033[1;31m●\033[0m"
	case OverlaySuccess:
		prefix = "[sussurai] \033[1;32m✓\033[0m"
	case OverlayError:
		prefix = "[sussurai] \033[1;31m✗\033[0m"
	case OverlayTranscribing:
		prefix = "[sussurai] \033[1;33m◌\033[0m"
	}
	fmt.Printf("%s %s\n", prefix, text)
}
