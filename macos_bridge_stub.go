//go:build !darwin

package main

type WindowInfo struct {
	OwnerName string  `json:"ownerName"`
	Title     string  `json:"title"`
	Icon      string  `json:"icon"`
	PID       int     `json:"pid"`
	Layer     int     `json:"layer"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
}

type PermissionStatus struct {
	Accessibility    bool     `json:"accessibility"`
	ScreenRecording  bool     `json:"screenRecording"`
	HotkeyRegistered bool     `json:"hotkeyRegistered"`
	HotkeyPresses    int      `json:"hotkeyPresses"`
	StatusItemReady  bool     `json:"statusItemReady"`
	Warnings         []string `json:"warnings"`
}

func getPermissionStatus(hotkeyRegistered bool, hotkeyPresses int, statusItemReady bool) PermissionStatus {
	return PermissionStatus{
		HotkeyRegistered: hotkeyRegistered,
		HotkeyPresses:    hotkeyPresses,
		StatusItemReady:  statusItemReady,
		Warnings:         []string{"macOS-only native integration is unavailable on this platform."},
	}
}

func requestAccessibilityPermission() bool {
	return false
}

func accessibilityTrusted() bool {
	return false
}

func requestScreenRecordingPermission() bool {
	return false
}

func listWindows() ([]WindowInfo, error) {
	return []WindowInfo{}, nil
}

func activateApp(pid int) bool {
	return false
}

func activateWindow(pid int, title string) bool {
	return false
}

func frontmostAppPID() int {
	return 0
}

func registerOptionTabHotkey() bool {
	return false
}

func consumeOptionTabPressCount() int {
	return 0
}

func consumeOptionKeyReleaseCount() int {
	return 0
}

func registerStatusItem() bool {
	return false
}

func consumeStatusAction() int {
	return 0
}

func configureWindowAppearance() {}
