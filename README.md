# Fast Switch

Minimal Wails scaffold for a macOS app that customizes Alt-Tab style switching.

## Stack

- Go backend with Wails
- React + TypeScript frontend
- Intended macOS-specific integrations added in later steps

## Run

1. Install the Wails CLI:
   `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
2. Install frontend dependencies:
   `cd frontend && npm install`
3. Start the app in development mode:
   `wails dev`

## Current status

This scaffold provides:

- A working desktop app shell
- A frontend landing screen oriented around the switcher concept
- A simple Go API (`GetAppState`) ready for frontend integration

It does not yet implement:

- Global hotkeys
- Window enumeration
- App/window focus switching
- Accessibility permission handling
