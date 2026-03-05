# Contributing to Sussurai

Thanks for your interest in contributing!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/sussurai.git`
3. Build: `make`
4. Run tests: `make test`

## Development Setup

You'll need:
- Go 1.22+
- C/C++ compiler and CMake (for whisper.cpp)
- `wl-clipboard` or `xclip`
- Access to `/dev/uinput` and `/dev/input`

## Making Changes

1. Create a branch: `git checkout -b my-feature`
2. Make your changes
3. Run tests: `make test`
4. Commit with a clear message
5. Push and open a pull request

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Handle all errors explicitly

## Reporting Issues

- Include your Linux distro and desktop environment
- Include the terminal output from running `sussurai`
- Describe what you expected vs what happened
