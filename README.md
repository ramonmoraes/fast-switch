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
- A frontend that reads live desktop state from the Go backend
- Native macOS window enumeration through CoreGraphics
- Native app/window activation through Accessibility APIs
- Command+Tab interception through a macOS event tap
- A menu bar status item (`FS`) with switcher and refresh actions

It does not yet implement:

- Automatic dismissal on `Option` key release
- Ranking, recency, and search
- Window thumbnails or previews

## macOS permissions

To get useful results, expect to grant:

- Accessibility, for focusing specific windows
- Screen Recording, for fuller window metadata on modern macOS
