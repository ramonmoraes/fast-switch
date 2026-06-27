import { useEffect, useRef, useState } from "react";
import { EventsOn, WindowSetSize } from "../wailsjs/runtime/runtime";

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
  icon: string;
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
          ActivateWindow: (pid: number, title: string) => Promise<boolean>;
          ShowSwitcher: () => Promise<DesktopSnapshot>;
          MoveSelection: (direction: number) => Promise<DesktopSnapshot>;
          ConfirmSelection: () => Promise<boolean>;
          CancelSwitcher: () => Promise<void>;
        };
      };
    };
  }
}

const fallbackSnapshot: DesktopSnapshot = {
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

function iconSource(windowInfo: WindowInfo) {
  return windowInfo.icon ? `data:image/png;base64,${windowInfo.icon}` : "";
}

function iconFallback(windowInfo: WindowInfo) {
  return (windowInfo.ownerName || windowInfo.title || "?").slice(0, 1).toUpperCase();
}

function App() {
  const [snapshot, setSnapshot] = useState<DesktopSnapshot>(fallbackSnapshot);
  const switcherRef = useRef<HTMLElement | null>(null);

  async function refreshSnapshot() {
    if (!window.go?.main?.App?.GetDesktopSnapshot) {
      return;
    }
    const next = await window.go.main.App.GetDesktopSnapshot();
    setSnapshot(next);
  }

  useEffect(() => {
    void refreshSnapshot();

    const unsubscribe = EventsOn("switcher:snapshot", (next: DesktopSnapshot) => {
      setSnapshot(next);
    });

    return () => unsubscribe();
  }, []);

  useEffect(() => {
    const switcherElement = switcherRef.current;
    if (!switcherElement) {
      return;
    }
    const measuredElement: HTMLElement = switcherElement;

    function syncWindowSize() {
      const rect = measuredElement.getBoundingClientRect();
      WindowSetSize(Math.ceil(rect.width), Math.ceil(rect.height));
    }

    syncWindowSize();

    const observer = new ResizeObserver(() => {
      syncWindowSize();
    });
    observer.observe(measuredElement);

    return () => observer.disconnect();
  }, [snapshot.windows.length]);

  useEffect(() => {
    async function onKeyDown(event: KeyboardEvent) {
      const app = window.go?.main?.App;
      if (!app) {
        return;
      }

      if (event.key === "Tab" || event.key === "ArrowRight") {
        event.preventDefault();
        setSnapshot(await app.MoveSelection(1));
        return;
      }

      if (event.key === "ArrowLeft") {
        event.preventDefault();
        setSnapshot(await app.MoveSelection(-1));
        return;
      }

      if (event.key === "Enter") {
        event.preventDefault();
        await app.ConfirmSelection();
        return;
      }

      if (event.key === "Escape") {
        event.preventDefault();
        await app.CancelSwitcher();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const selectedWindow =
    snapshot.switcher.selectedWindow ??
    (snapshot.switcher.selectedIndex >= 0 ? snapshot.windows[snapshot.switcher.selectedIndex] : undefined);

  return (
    <main className="switcher-shell">
      {!snapshot.permissions.hotkeyRegistered && snapshot.permissions.accessibility ? (
        <div className="permission-note">Command-Tab interception unavailable</div>
      ) : null}

      {!snapshot.permissions.accessibility ? (
        <button
          type="button"
          className="permission-note permission-button"
          onClick={() => void window.go?.main?.App?.RequestAccessibilityPermission?.()}
        >
          Enable Accessibility To Use Fast Switch
        </button>
      ) : null}

      <section className="switcher-frame" aria-label="App Switcher" id="switcher" ref={switcherRef}>
        <div className="switcher-strip">
          {snapshot.windows.map((item, index) => (
            <button
              type="button"
              className={`app-tile${snapshot.switcher.selectedIndex === index ? " active" : ""}`}
              key={`${item.pid}-${item.ownerName}`}
              onClick={() => void window.go?.main?.App?.ActivateWindow?.(item.pid, item.title)}
            >
              <div className="app-icon-wrap">
                {item.icon ? (
                  <img className="app-icon" src={iconSource(item)} alt="" draggable={false} />
                ) : (
                  <div className="app-icon app-icon-fallback">{iconFallback(item)}</div>
                )}
              </div>
              <span className="app-label">{item.ownerName || item.title || "App"}</span>
            </button>
          ))}
        </div>
      </section>
    </main>
  );
}

export default App;
