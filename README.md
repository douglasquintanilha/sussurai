# Sussurai

Free, open-source push-to-talk voice-to-text for Linux. Hold a key, speak, release — your words are transcribed and pasted into any app.

A lightweight alternative to [Superwhisper](https://superwhisper.com/) and [Wispr Flow](https://wisprflow.com/) that runs natively on Linux with Wayland support.

## Features

- **Push-to-talk**: Hold Right Alt to record, release to transcribe and paste
- **Toggle mode**: Double-tap Right Alt for hands-free recording
- **Fast transcription**: Uses [Groq](https://groq.com/) (free) or [OpenAI](https://openai.com/) Whisper API, or local [whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- **Multi-language**: Auto-detects English, Portuguese, and 90+ other languages
- **Works everywhere**: Pastes into any focused app (terminals, browsers, editors)
- **System tray**: Minimal tray icon shows recording state
- **Wayland native**: Built for modern Linux desktops (COSMIC, GNOME, KDE)

## Quick Start

### Prerequisites

- Go 1.22+
- C/C++ compiler (gcc/g++)
- CMake
- `wl-copy` (Wayland) or `xclip` (X11)

```bash
# Ubuntu/Pop!_OS
sudo apt install build-essential cmake golang wl-clipboard

# Access to /dev/uinput (for pasting)
sudo setfacl -m u:$USER:rw /dev/uinput

# Access to /dev/input (for hotkey capture)
sudo usermod -aG input $USER
# Log out and back in for group change to take effect
```

### Build & Install

```bash
git clone https://github.com/raphaelfp/sussurai.git
cd sussurai
make              # Clones whisper.cpp + builds everything
make install      # Installs to ~/.local/bin, adds to app menu & autostart
```

### Configure

```bash
# Copy the example env file and add your API key
cp .env.example ~/.config/sussurai/.env
# Edit ~/.config/sussurai/.env and add your Groq key (free at https://console.groq.com)
```

Or create `~/.config/sussurai/config.toml`:

```toml
hotkey = "RightAlt"
backend = "groq"           # "groq", "openai", or "local"
double_tap_ms = 300

[groq]
api_key = "gsk_..."        # or set GROQ_API_KEY env var

[openai]
api_key = "sk-..."         # or set OPENAI_API_KEY env var

[local]
model_size = "base"        # tiny, base, small, medium, large
language = "auto"           # auto, en, pt, es, fr, ...
```

### Translation Mode

To enable translation mode, set `translate = true` under the `[groq]` or `[openai]` section in your `~/.config/sussurai/config.toml`:

```toml
[groq]
api_key = "gsk_..."
translate = true

[openai]
api_key = "sk-..."
translate = true
```

In this mode, Sussurai uses the translations endpoint instead of the transcriptions endpoint — it auto-detects any input language and always outputs English.

Default: disabled (standard transcription, preserves input language).

### Vocabulary File

To improve transcription accuracy for specific words or phrases, create a vocabulary file at `~/.config/sussurai/vocabulary.txt`:

```
# Sussurai vocabulary — loaded on every transcribe call
# Add words/phrases that get misheard, one per line
# Lines starting with # are comments
Claude
Claude Code
CLAUDE.md
Sussurai
```

The file is loaded on every transcribe call, so you can edit it without restarting Sussurai. Words are sent as a prompt hint to the Whisper API (works with both Groq and OpenAI backends).

### Run

```bash
sussurai
# Or just find "Sussurai" in your app menu
```

## Usage

| Action | What happens |
|---|---|
| **Hold** Right Alt + speak + **release** | Records → transcribes → pastes text |
| **Double-tap** Right Alt + speak + **tap** again | Toggle mode: same result, hands-free |

The tray icon changes shape to show state:
- Small dot = idle
- Large dot = recording
- Ring = transcribing

## Backends

| Backend | Quality | Speed | Cost |
|---|---|---|---|
| **Groq** (recommended) | Excellent (whisper-large-v3) | Very fast | Free tier available |
| **OpenAI** | Excellent (whisper-1) | Fast | ~$0.006/min |
| **Local** | Varies by model size | Depends on hardware | Free |

## Supported Hotkeys

`RightAlt` (default), `LeftAlt`, `RightCtrl`, `LeftCtrl`, `RightShift`, `LeftShift`, `RightMeta`, `LeftMeta`, `CapsLock`, `ScrollLock`, `Pause`, `F13`, `F14`, `F15`

## Uninstall

```bash
make uninstall
```

## License

[MIT](LICENSE)
