package main

import (
	"context"
	"runtime"
)

type App struct {
	ctx context.Context
}

type AppState struct {
	Name         string   `json:"name"`
	Platform     string   `json:"platform"`
	Capabilities []string `json:"capabilities"`
	NextSteps    []string `json:"nextSteps"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetAppState() AppState {
	return AppState{
		Name:     "Fast Switch",
		Platform: runtime.GOOS,
		Capabilities: []string{
			"Desktop shell with Go backend",
			"Custom switcher overlay UI",
			"Room for macOS hotkey and window integration",
		},
		NextSteps: []string{
			"Add global hotkey support for Option+Tab",
			"Bridge macOS accessibility/window APIs",
			"Replace mock data with live window state",
		},
	}
}
