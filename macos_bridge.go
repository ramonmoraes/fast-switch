//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices -framework Carbon
#include <stdbool.h>
#include <stdlib.h>

char* fastswitch_copy_windows_json(void);
bool fastswitch_accessibility_trusted(void);
bool fastswitch_request_accessibility(void);
bool fastswitch_screen_recording_granted(void);
bool fastswitch_request_screen_recording(void);
bool fastswitch_activate_app(int pid);
bool fastswitch_activate_window(int pid, const char* title);
int fastswitch_frontmost_app_pid(void);
bool fastswitch_register_option_tab_hotkey(void);
int fastswitch_consume_option_tab_press_count(void);
int fastswitch_consume_option_key_release_count(void);
bool fastswitch_register_status_item(void);
int fastswitch_consume_status_action(void);
void fastswitch_configure_window_appearance(void);
void fastswitch_free_string(char* value);
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

type WindowInfo struct {
	OwnerName string  `json:"ownerName"`
	Title     string  `json:"title"`
	Icon      string  `json:"icon"`
	PID       int     `json:"pid"`
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
	status := PermissionStatus{
		Accessibility:    bool(C.fastswitch_accessibility_trusted()),
		ScreenRecording:  bool(C.fastswitch_screen_recording_granted()),
		HotkeyRegistered: hotkeyRegistered,
		HotkeyPresses:    hotkeyPresses,
		StatusItemReady:  statusItemReady,
	}

	if !status.Accessibility {
		status.Warnings = append(status.Warnings, "Accessibility access is required to focus specific windows.")
	}
	if !status.ScreenRecording {
		status.Warnings = append(status.Warnings, "Screen Recording may be required for complete window metadata on macOS.")
	}
	if !status.HotkeyRegistered {
		status.Warnings = append(status.Warnings, "Command+Tab interception is not active yet.")
	}
	if !status.StatusItemReady {
		status.Warnings = append(status.Warnings, "Menu bar status item is not active yet.")
	}

	return status
}

func requestAccessibilityPermission() bool {
	return bool(C.fastswitch_request_accessibility())
}

func accessibilityTrusted() bool {
	return bool(C.fastswitch_accessibility_trusted())
}

func requestScreenRecordingPermission() bool {
	return bool(C.fastswitch_request_screen_recording())
}

func listWindows() ([]WindowInfo, error) {
	jsonString := C.fastswitch_copy_windows_json()
	if jsonString == nil {
		return []WindowInfo{}, nil
	}
	defer C.fastswitch_free_string(jsonString)

	raw := C.GoString(jsonString)
	if raw == "" {
		return []WindowInfo{}, nil
	}

	var windows []WindowInfo
	if err := json.Unmarshal([]byte(raw), &windows); err != nil {
		return nil, err
	}
	return windows, nil
}

func activateApp(pid int) bool {
	return bool(C.fastswitch_activate_app(C.int(pid)))
}

func activateWindow(pid int, title string) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.fastswitch_activate_window(C.int(pid), cTitle))
}

func frontmostAppPID() int {
	return int(C.fastswitch_frontmost_app_pid())
}

func registerOptionTabHotkey() bool {
	return bool(C.fastswitch_register_option_tab_hotkey())
}

func consumeOptionTabPressCount() int {
	return int(C.fastswitch_consume_option_tab_press_count())
}

func consumeOptionKeyReleaseCount() int {
	return int(C.fastswitch_consume_option_key_release_count())
}

func registerStatusItem() bool {
	return bool(C.fastswitch_register_status_item())
}

func consumeStatusAction() int {
	return int(C.fastswitch_consume_status_action())
}

func configureWindowAppearance() {
	C.fastswitch_configure_window_appearance()
}
