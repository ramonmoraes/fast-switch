const mockWindows = [
  { app: "Cursor", title: "main.go", status: "Focused 2m ago" },
  { app: "Safari", title: "Wails docs", status: "Pinned for setup" },
  { app: "Terminal", title: "wails dev", status: "Background session" },
  { app: "Figma", title: "Switcher layout", status: "Recent design" },
];

function App() {
  return (
    <main className="app-shell">
      <section className="hero">
        <p className="eyebrow">macOS Alt-Tab Utility</p>
        <h1>Fast Switch</h1>
        <p className="lede">
          A Wails-based desktop shell for a customizable app and window switcher.
          The frontend is ready for live data once the macOS integration layer is added.
        </p>
      </section>

      <section className="panel">
        <div className="panel-header">
          <div>
            <p className="section-label">Overlay Preview</p>
            <h2>Switcher concept</h2>
          </div>
          <span className="shortcut">Option + Tab</span>
        </div>

        <div className="switcher">
          {mockWindows.map((item) => (
            <article className="switcher-card" key={`${item.app}-${item.title}`}>
              <div className="switcher-icon">{item.app.slice(0, 1)}</div>
              <div className="switcher-copy">
                <strong>{item.app}</strong>
                <span>{item.title}</span>
                <small>{item.status}</small>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="grid">
        <article className="info-card">
          <p className="section-label">Build track</p>
          <h3>Why Wails</h3>
          <p>
            Go handles app logic and future native bridges, while the UI stays fast to
            iterate with React.
          </p>
        </article>
        <article className="info-card">
          <p className="section-label">Next milestone</p>
          <h3>Native integration</h3>
          <p>
            Add macOS-specific hotkey capture and window access through a focused native
            bridge instead of pushing everything into the web layer.
          </p>
        </article>
      </section>
    </main>
  );
}

export default App;
