package main

import (
	"context"
	goruntime "runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	statusActionNone         = 0
	statusActionOpenSwitcher = 1
	statusActionRefresh      = 2
	statusActionQuit         = 3
	switcherMinWidth         = 340
	switcherMaxWidth         = 1120
	switcherHeight           = 214
	switcherTileWidth        = 92
	switcherFramePadding     = 64
)

type App struct {
	ctx              context.Context
	mu               sync.Mutex
	hotkeyRegistered bool
	statusItemReady  bool
	hotkeyPresses    int
	windows          []WindowInfo
	selectedIndex    int
	switcherVisible  bool
}

type AppState struct {
	Name         string   `json:"name"`
	Platform     string   `json:"platform"`
	Capabilities []string `json:"capabilities"`
	NextSteps    []string `json:"nextSteps"`
}

type SwitcherState struct {
	Visible        bool        `json:"visible"`
	SelectedIndex  int         `json:"selectedIndex"`
	SelectedWindow *WindowInfo `json:"selectedWindow,omitempty"`
}

type DesktopSnapshot struct {
	AppState    AppState         `json:"appState"`
	Permissions PermissionStatus `json:"permissions"`
	Windows     []WindowInfo     `json:"windows"`
	Switcher    SwitcherState    `json:"switcher"`
}

func NewApp() *App {
	return &App{
		selectedIndex: -1,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.hotkeyRegistered = registerOptionTabHotkey()
	a.statusItemReady = registerStatusItem()
	wruntime.WindowSetAlwaysOnTop(ctx, true)
	wruntime.WindowCenter(ctx)
	go a.pollNativeInputs()
}

func (a *App) GetAppState() AppState {
	return AppState{
		Name:     "Fast Switch",
		Platform: goruntime.GOOS,
		Capabilities: []string{
			"Desktop shell with Go backend",
			"Transient switcher overlay window",
			"Live macOS permissions, menu bar, and activation",
		},
		NextSteps: []string{
			"Dismiss the switcher automatically on key release",
			"Add ranking, search, and recency ordering",
			"Use native previews instead of metadata-only cards",
		},
	}
}

func (a *App) GetDesktopSnapshot() DesktopSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureNativeHooksLocked()
	a.refreshWindowsLocked()
	return a.snapshotLocked()
}

func (a *App) RequestAccessibilityPermission() PermissionStatus {
	requestAccessibilityPermission()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureNativeHooksLocked()
	return a.permissionStatusLocked()
}

func (a *App) RequestScreenRecordingPermission() PermissionStatus {
	requestScreenRecordingPermission()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureNativeHooksLocked()
	return a.permissionStatusLocked()
}

func (a *App) ActivateWindow(pid int, title string) bool {
	activated := activateWindow(pid, title)
	if !activated {
		activated = activateApp(pid)
	}
	if activated {
		a.hideSwitcher()
	}
	return activated
}

func (a *App) ShowSwitcher() DesktopSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.showSwitcherLocked()
	return a.snapshotLocked()
}

func (a *App) MoveSelection(direction int) DesktopSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.refreshWindowsLocked()
	a.moveSelectionLocked(direction)
	return a.snapshotLocked()
}

func (a *App) ConfirmSelection() bool {
	a.mu.Lock()
	if len(a.windows) == 0 || a.selectedIndex < 0 || a.selectedIndex >= len(a.windows) {
		a.mu.Unlock()
		a.hideSwitcher()
		return false
	}
	selected := a.windows[a.selectedIndex]
	a.mu.Unlock()
	return a.ActivateWindow(selected.PID, selected.Title)
}

func (a *App) CancelSwitcher() {
	a.hideSwitcher()
}

func (a *App) RefreshWindows() DesktopSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.refreshWindowsLocked()
	return a.snapshotLocked()
}

func (a *App) pollNativeInputs() {
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		a.ensureNativeHooksLocked()
		a.mu.Unlock()

		pressCount := consumeOptionTabPressCount()
		if pressCount > 0 {
			a.handleHotkeyPresses(pressCount)
		}

		releaseCount := consumeOptionKeyReleaseCount()
		if releaseCount > 0 {
			a.handleOptionRelease()
		}

		switch consumeStatusAction() {
		case statusActionOpenSwitcher:
			a.handleOpenSwitcherAction()
		case statusActionRefresh:
			a.handleRefreshAction()
		case statusActionQuit:
			wruntime.Quit(a.ctx)
			return
		}
	}
}

func (a *App) ensureNativeHooksLocked() {
	if !a.statusItemReady {
		a.statusItemReady = registerStatusItem()
	}
	if !a.hotkeyRegistered && accessibilityTrusted() {
		a.hotkeyRegistered = registerOptionTabHotkey()
	}
}

func (a *App) handleOptionRelease() {
	a.mu.Lock()
	visible := a.switcherVisible
	hasSelection := len(a.windows) > 0 && a.selectedIndex >= 0 && a.selectedIndex < len(a.windows)
	selected := WindowInfo{}
	if hasSelection {
		selected = a.windows[a.selectedIndex]
	}
	a.mu.Unlock()

	if !visible {
		return
	}

	if hasSelection {
		_ = a.ActivateWindow(selected.PID, selected.Title)
		return
	}

	a.hideSwitcher()
}

func (a *App) handleHotkeyPresses(count int) {
	a.mu.Lock()
	a.hotkeyPresses += count
	a.refreshWindowsLocked()
	frontmostPID := frontmostAppPID()
	if len(a.windows) == 0 {
		a.switcherVisible = true
		a.selectedIndex = -1
	} else if !a.switcherVisible {
		a.switcherVisible = true
		a.selectedIndex = a.firstSelectableIndexLocked(frontmostPID)
		if a.selectedIndex < 0 {
			a.selectedIndex = 0
		}
		if count > 1 && len(a.windows) > 0 {
			a.selectedIndex = (a.selectedIndex + (count - 1)) % len(a.windows)
		}
	} else {
		a.selectedIndex = (a.selectedIndex + count) % len(a.windows)
	}
	snapshot := a.snapshotLocked()
	a.mu.Unlock()
	a.presentSnapshot(snapshot)
}

func (a *App) handleOpenSwitcherAction() {
	a.mu.Lock()
	a.showSwitcherLocked()
	snapshot := a.snapshotLocked()
	a.mu.Unlock()
	a.presentSnapshot(snapshot)
}

func (a *App) handleRefreshAction() {
	a.mu.Lock()
	a.refreshWindowsLocked()
	snapshot := a.snapshotLocked()
	a.mu.Unlock()
	a.emitSnapshot(snapshot)
}

func (a *App) showSwitcherLocked() {
	a.refreshWindowsLocked()
	a.switcherVisible = true
	if len(a.windows) == 0 {
		a.selectedIndex = -1
		return
	}
	if a.selectedIndex < 0 || a.selectedIndex >= len(a.windows) {
		a.selectedIndex = 0
	}
}

func (a *App) moveSelectionLocked(direction int) {
	if len(a.windows) == 0 {
		a.selectedIndex = -1
		return
	}
	if !a.switcherVisible {
		a.switcherVisible = true
	}
	if a.selectedIndex < 0 || a.selectedIndex >= len(a.windows) {
		a.selectedIndex = 0
		return
	}
	mod := len(a.windows)
	a.selectedIndex = (a.selectedIndex + direction + mod) % mod
}

func (a *App) refreshWindowsLocked() {
	windows, err := listWindows()
	if err != nil {
		a.windows = []WindowInfo{}
		a.selectedIndex = -1
		return
	}

	windows = collapseToPrimaryWindows(windows)
	currentSelection := a.selectedWindowKeyLocked()
	a.windows = windows
	if len(a.windows) == 0 {
		a.selectedIndex = -1
		return
	}
	if currentSelection != "" {
		for index, window := range a.windows {
			if windowKey(window) == currentSelection {
				a.selectedIndex = index
				return
			}
		}
	}
	if a.selectedIndex < 0 || a.selectedIndex >= len(a.windows) {
		a.selectedIndex = 0
	}
}

func (a *App) selectedWindowKeyLocked() string {
	if a.selectedIndex < 0 || a.selectedIndex >= len(a.windows) {
		return ""
	}
	return windowKey(a.windows[a.selectedIndex])
}

func windowKey(window WindowInfo) string {
	return window.OwnerName + "|" + window.Title + "|" + strconv.Itoa(window.PID)
}

func (a *App) firstSelectableIndexLocked(frontmostPID int) int {
	if len(a.windows) == 0 {
		return -1
	}
	for index, window := range a.windows {
		if window.PID != frontmostPID {
			return index
		}
	}
	return 0
}

func collapseToPrimaryWindows(windows []WindowInfo) []WindowInfo {
	type scoredWindow struct {
		window WindowInfo
		score  float64
	}

	bestByApp := make(map[int]scoredWindow)
	order := make([]int, 0, len(windows))

	for _, window := range windows {
		score := window.Width * window.Height
		if window.Title != "" {
			score += 1000000
		}

		current, exists := bestByApp[window.PID]
		if !exists {
			bestByApp[window.PID] = scoredWindow{window: window, score: score}
			order = append(order, window.PID)
			continue
		}

		if score > current.score {
			bestByApp[window.PID] = scoredWindow{window: window, score: score}
		}
	}

	result := make([]WindowInfo, 0, len(bestByApp))
	for _, pid := range order {
		if selected, exists := bestByApp[pid]; exists {
			result = append(result, selected.window)
			delete(bestByApp, pid)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		leftArea := result[i].Width * result[i].Height
		rightArea := result[j].Width * result[j].Height
		if leftArea == rightArea {
			return result[i].OwnerName < result[j].OwnerName
		}
		return leftArea > rightArea
	})

	return result
}

func (a *App) snapshotLocked() DesktopSnapshot {
	windows := append([]WindowInfo(nil), a.windows...)
	switcher := SwitcherState{
		Visible:       a.switcherVisible,
		SelectedIndex: a.selectedIndex,
	}
	if a.selectedIndex >= 0 && a.selectedIndex < len(windows) {
		selected := windows[a.selectedIndex]
		switcher.SelectedWindow = &selected
	}
	return DesktopSnapshot{
		AppState:    a.GetAppState(),
		Permissions: a.permissionStatusLocked(),
		Windows:     windows,
		Switcher:    switcher,
	}
}

func (a *App) permissionStatusLocked() PermissionStatus {
	return getPermissionStatus(a.hotkeyRegistered, a.hotkeyPresses, a.statusItemReady)
}

func (a *App) presentSnapshot(snapshot DesktopSnapshot) {
	wruntime.WindowSetSize(a.ctx, calculateSwitcherWidth(len(snapshot.Windows)), switcherHeight)
	wruntime.WindowCenter(a.ctx)
	wruntime.WindowShow(a.ctx)
	wruntime.Show(a.ctx)
	a.emitSnapshot(snapshot)
}

func (a *App) emitSnapshot(snapshot DesktopSnapshot) {
	wruntime.EventsEmit(a.ctx, "switcher:snapshot", snapshot)
}

func (a *App) hideSwitcher() {
	a.mu.Lock()
	a.switcherVisible = false
	snapshot := a.snapshotLocked()
	a.mu.Unlock()
	wruntime.WindowHide(a.ctx)
	wruntime.Hide(a.ctx)
	a.emitSnapshot(snapshot)
}

func calculateSwitcherWidth(windowCount int) int {
	if windowCount <= 0 {
		return switcherMinWidth
	}
	width := switcherFramePadding + (windowCount * switcherTileWidth)
	if width < switcherMinWidth {
		return switcherMinWidth
	}
	if width > switcherMaxWidth {
		return switcherMaxWidth
	}
	return width
}
