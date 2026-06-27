import { useEffect, useState } from "react";

type AppState = {
  name: string;
  platform: string;
  capabilities: string[];
  nextSteps: string[];
};

type PermissionStatus = {
  accessibility: boolean;
  screenRecording: boolean;
  hotkeyRegistered: boolean;
  hotkeyPresses: number;
  statusItemReady: boolean;
  warnings: string[];
};

type WindowInfo = {
  ownerName: string;
  title: string;
  pid: number;
  layer: number;
  x: number;
  y: number;
  width: number;
  height: number;
};

type SwitcherState = {
  visible: boolean;
  selectedIndex: number;
  selectedWindow?: WindowInfo;
};

type DesktopSnapshot = {
  appState: AppState;
  permissions: PermissionStatus;
  windows: WindowInfo[];
  switcher: SwitcherState;
};

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetDesktopSnapshot: () => Promise<DesktopSnapshot>;
          RequestAccessibilityPermission: () => Promise<PermissionStatus>;
          RequestScreenRecordingPermission: () => Promise<PermissionStatus>;
          ActivateWindow: (pid: number, title: string) => Promise<boolean>;
          ShowSwitcher: () => Promise<DesktopSnapshot>;
          MoveSelection: (direction: number) => Promise<DesktopSnapshot>;
          ConfirmSelection: () => Promise<boolean>;
          CancelSwitcher: () => Promise<void>;
          RefreshWindows: () => Promise<DesktopSnapshot>;
        };
      };
    };
  }
}

const fallbackSnapshot: DesktopSnapshot = {
  appState: {
    name: "Fast Switch",
    platform: "darwin",
    capabilities: [
      "Desktop shell with Go backend",
      "Transient switcher overlay window",
      "Live macOS permissions, menu bar, and activation",
    ],
    nextSteps: [
      "Dismiss the switcher automatically on key release",
      "Add ranking, search, and recency ordering",
      "Use native previews instead of metadata-only cards",
    ],
  },
  permissions: {
    accessibility: false,
    screenRecording: false,
    hotkeyRegistered: false,
    hotkeyPresses: 0,
    statusItemReady: false,
    warnings: ["Run this UI inside Wails to access live macOS data."],
  },
  windows: [],
  switcher: {
    visible: true,
    selectedIndex: -1,
  },
};

function App() {
  const [snapshot, setSnapshot] = useState<DesktopSnapshot>(fallbackSnapshot);
  const [loading, setLoading] = useState(true);
  const [activationMessage, setActivationMessage] = useState("");

  async function refreshSnapshot() {
    if (!window.go?.main?.App?.GetDesktopSnapshot) {
      setLoading(false);
      return;
    }

    const next = await window.go.main.App.GetDesktopSnapshot();
    setSnapshot(next);
    setLoading(false);
  }

  useEffect(() => {
    void refreshSnapshot();

    const timer = window.setInterval(() => {
      void refreshSnapshot();
    }, 160);

    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    async function onKeyDown(event: KeyboardEvent) {
      if (!window.go?.main?.App) {
        return;
      }

      if (!snapshot.switcher.visible) {
        if (event.key === "Enter") {
          event.preventDefault();
          const next = await window.go.main.App.ShowSwitcher();
          setSnapshot(next);
        }
        return;
      }

      if (event.key === "Tab" || event.key === "ArrowRight" || event.key === "ArrowDown") {
        event.preventDefault();
        const next = await window.go.main.App.MoveSelection(1);
        setSnapshot(next);
        return;
      }

      if (event.key === "ArrowLeft" || event.key === "ArrowUp") {
        event.preventDefault();
        const next = await window.go.main.App.MoveSelection(-1);
        setSnapshot(next);
        return;
      }

      if (event.key === "Enter") {
        event.preventDefault();
        const activated = await window.go.main.App.ConfirmSelection();
        setActivationMessage(activated ? "Activated selected window" : "Unable to activate selected window");
        return;
      }

      if (event.key === "Escape") {
        event.preventDefault();
        await window.go.main.App.CancelSwitcher();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [snapshot.switcher.visible]);

  async function requestAccessibility() {
    const next = await window.go?.main?.App?.RequestAccessibilityPermission?.();
    if (next) {
      setSnapshot((current) => ({ ...current, permissions: next }));
    }
  }

  async function requestScreenRecording() {
    const next = await window.go?.main?.App?.RequestScreenRecordingPermission?.();
    if (next) {
      setSnapshot((current) => ({ ...current, permissions: next }));
    }
  }

  async function refreshWindows() {
    const next = await window.go?.main?.App?.RefreshWindows?.();
    if (next) {
      setSnapshot(next);
    }
  }

  async function activate(windowInfo: WindowInfo) {
    const activated = await window.go?.main?.App?.ActivateWindow?.(windowInfo.pid, windowInfo.title);
    setActivationMessage(
      activated
        ? `Activated ${windowInfo.ownerName}${windowInfo.title ? `: ${windowInfo.title}` : ""}`
        : `Unable to activate ${windowInfo.ownerName}${windowInfo.title ? `: ${windowInfo.title}` : ""}`,
    );
    void refreshSnapshot();
  }

  const hotkeyText = snapshot.permissions.hotkeyRegistered
    ? `Command + Tab live (${snapshot.permissions.hotkeyPresses} presses)`
    : "Command + Tab not registered";

  const debug = false

  return (
    <main className={`app-shell${snapshot.switcher.visible ? " switcher-mode" : ""}`}>
      <section className="panel">
        {debug && (
          <div className="toolbar">
            <button type="button" onClick={() => void refreshWindows()}>
              Refresh windows
            </button>
            <button type="button" onClick={() => void requestAccessibility()}>
              Request Accessibility
            </button>
            <button type="button" onClick={() => void requestScreenRecording()}>
              Request Screen Recording
            </button>
          </div>
        )}

        {debug && (
          <div className="status-grid">
            <article className="status-card">
              <span>Accessibility</span>
              <strong>{snapshot.permissions.accessibility ? "Granted" : "Missing"}</strong>
            </article>
            <article className="status-card">
              <span>Screen Recording</span>
              <strong>{snapshot.permissions.screenRecording ? "Granted" : "Missing"}</strong>
            </article>
            <article className="status-card">
              <span>Menu Bar</span>
              <strong>{snapshot.permissions.statusItemReady ? "Registered" : "Unavailable"}</strong>
            </article>
            <article className="status-card">
              <span>Switcher</span>
              <strong>{snapshot.switcher.visible ? "Visible" : "Hidden"}</strong>
            </article>
          </div>
        )}

        {debug && snapshot.permissions.warnings.length > 0 ? (
          <div className="warning-list">
            {snapshot.permissions.warnings.map((warning) => (
              <p key={warning}>{warning}</p>
            ))}
          </div>
        ) : null}

        {activationMessage ? <p className="activation-message">{activationMessage}</p> : null}

        <div className="switcher">
          {snapshot.windows.map((item, index) => (
            <article
              className={`switcher-card${snapshot.switcher.selectedIndex === index ? " active" : ""}`}
              key={`${item.pid}-${item.ownerName}-${item.title}`}
            >
              <div className="switcher-icon">{item.ownerName.slice(0, 1)}</div>
              <div className="switcher-copy">
                <strong>{item.ownerName || item.title || "Untitled window"}</strong>
              </div>

              {debug && (
                <button type="button" className="activate-button" onClick={() => void activate(item)}>
                  Activate
                </button>
              )}
            </article>
          ))}
        </div>

        {!loading && snapshot.windows.length === 0 ? (
          <p className="empty-state">
            No switchable windows were found. If this is unexpected, grant permissions and refresh.
          </p>
        ) : null}
      </section>
    </main>
  );
}

export default App;
