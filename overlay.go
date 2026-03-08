package main

import (
	"fmt"
	"sync/atomic"
)

var overlayGen atomic.Uint64

type OverlayState int

const (
	OverlayIdle         OverlayState = 0
	OverlayRecording    OverlayState = 1
	OverlayTranscribing OverlayState = 2
	OverlaySuccess      OverlayState = 3
	OverlayError        OverlayState = 4
)

func SetOverlay(state OverlayState, text string) {
	overlayGen.Add(1)
	SetTrayState(state)
	if text == "" {
		return
	}
	prefix := "[sussur.ai]"
	switch state {
	case OverlayRecording:
		prefix = "[sussur.ai] \033[1;31m●\033[0m"
	case OverlaySuccess:
		prefix = "[sussur.ai] \033[1;32m✓\033[0m"
	case OverlayError:
		prefix = "[sussur.ai] \033[1;31m✗\033[0m"
	case OverlayTranscribing:
		prefix = "[sussur.ai] \033[1;33m◌\033[0m"
	}
	fmt.Printf("%s %s\n", prefix, text)
}
