# Contributing to Sussur.ai

Thanks for your interest in contributing!

## Prerequisites

- Go 1.22+
- C/C++ compiler (gcc/g++) and CMake (for building whisper.cpp)
- `wl-clipboard` (Wayland) or `xclip` (X11)
- Access to `/dev/uinput` and `/dev/input` (see README for setup)

## Getting Started

```bash
# Fork the repo on GitHub, then:
git clone https://github.com/YOUR_USERNAME/sussurai.git
cd sussurai
make          # clones whisper.cpp + builds everything
make test     # run unit tests
```

## Making Changes

1. Create a branch: `git checkout -b my-feature`
2. Make your changes
3. Format and lint: `gofmt -w .` and `go vet ./...`
4. Run tests: `make test`
5. Commit with a clear message
6. Push and open a pull request

## Code Style

- Run `gofmt` and `go vet` before committing
- Keep functions focused and small
- Handle all errors explicitly — no `_` for error returns

## Reporting Issues

- Include your Linux distro, desktop environment, and display server (Wayland/X11)
- Include the terminal output from running `sussurai`
- Describe what you expected vs what happened
