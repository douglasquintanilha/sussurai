# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make                # Build whisper.cpp (if needed) + sussurai binary
make test           # Run all tests (sets CGo env vars automatically)
make install        # Install binary, desktop file, icon, autostart entry
make download-model # Download ggml-base.bin (~142MB) for local backend
make clean          # Remove binary and whisper.cpp build artifacts
```

Run a single test:
```bash
C_INCLUDE_PATH=$(pwd)/third_party/whisper.cpp/include:$(pwd)/third_party/whisper.cpp/ggml/include \
LIBRARY_PATH=$(pwd)/third_party/whisper.cpp/build/src:$(pwd)/third_party/whisper.cpp/build/ggml/src \
go test -run TestName -v ./...
```

`go vet` and `go build` also require those CGo env vars (or use `make` which sets them).

## Architecture

Sussur.ai is a Linux push-to-talk voice-to-text tool. Single Go package, CGo for whisper.cpp and miniaudio.

**Pipeline**: Hotkey press → Audio capture → Transcribe → Clipboard copy → Simulate Ctrl+Shift+V paste

**Event loop** (`main.go`): Listens for `KeyPress`/`KeyRelease` events from a channel, drives the pipeline sequentially. Global `activeTranscriber` and `appCfg` are protected by `sync.RWMutex` for runtime switching from the tray menu.

**Transcriber interface** (`transcribe.go`): `Transcribe([]float32) (string, error)` + `Close()`. Three backends:
- `groq.go` / `openai.go` — both use `apiTranscriber` (encodes WAV, multipart POST, parses JSON response)
- `local.go` — whisper.cpp via CGo bindings

**Runtime config switching** (`main.go:updateConfig`): Atomically creates a new transcriber from modified config, swaps it in under write lock, closes old one, saves config to TOML, refreshes tray menu checkmarks.

**Paste mechanism** (`paste.go`): Two steps — (1) `wl-copy`/`xclip` to clipboard, (2) pre-warmed uinput virtual keyboard simulates Ctrl+Shift+V. The keyboard is created at startup with a 2-second delay for compositor registration.

**Tray UI** (`tray.go`): fyne.io/systray with embedded PNG icons. Submenus for Backend, Language, Translate toggle, History (click-to-paste), Edit Vocabulary, Quit. History items paste with a 200ms delay to let the menu close.

**Input capture** (`input.go`): Reads all `/dev/input/event*` devices via evdev. Supports hold-to-record and double-tap toggle mode.

## Key Conventions

- App display name is **Sussur.ai** (binary/paths stay `sussurai`)
- Console log prefix: `[sussur.ai]`
- Config lives at `~/.config/sussurai/config.toml`, API keys in `.env` (never saved to TOML)
- `make install` preserves existing `~/.config/sussurai/.env` (won't overwrite)
- Vocabulary file (`vocabulary.txt`) is sent as the `prompt` field to API backends
- History persisted as flat JSON array at `~/.config/sussurai/history.json`

## Runtime Dependencies

- `wl-copy` (Wayland) or `xclip` (X11)
- `/dev/uinput` writable: `sudo setfacl -m u:$USER:rw /dev/uinput`
- User in `input` group for evdev access
