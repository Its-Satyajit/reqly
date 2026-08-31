import { useState, useEffect } from "react";
import { Settings } from "lucide-react";
import { useThemeStore } from "#stores";
import { THEMES } from "#lib/themes";
import { APP_VERSION } from "#lib/crash";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";
import { PageHeader } from "#components/PageHeader";
import { ProxyTlsPanel } from "./ProxyTlsPanel";
import { CicdPanel } from "./CicdPanel";

const SHORTCUTS: [string, string][] = [
  ["⌘K", "Toggle Command Palette"],
  ["⌘B", "Toggle Workspace Sidebar"],
  ["⌘J", "Toggle Bottom Panel (Console/Network)"],
  ["⌘W", "Close Active Tab (guarded against unsaved edits)"],
  ["⌘1–9", "Switch Tool View (Home, Requests, Envs, History, Mocks, Diff, JWT, GraphQL, Runners)"],
  ["⌘⏎", "Send Active Request"],
];

const RETENTION_OPTIONS = [
  { value: "30d", label: "30 days" },
  { value: "90d", label: "90 days (default)" },
  { value: "1yr", label: "1 year" },
  { value: "forever", label: "Forever" },
] as const;

export function SettingsView() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const workspace = useWorkspaceStore((s) => s.currentWorkspace);
  const pool = useHistoryStore((s) => s.pool);
  const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
  const [retention, setRetention] = useState("90d");
  useEffect(() => {
    const v = localStorage.getItem("reqly:historyRetention");
    if (v) setRetention(v);
  }, []);
  const saveRetention = (v: string) => {
    setRetention(v);
    localStorage.setItem("reqly:historyRetention", v);
  };
  return (
    <div className="flex h-full min-h-0 flex-col overflow-y-auto bg-background" aria-label="Settings">
      <PageHeader
        icon={Settings}
        title="Settings"
        description="Preferences, workspace configurations, proxy/TLS rules, and CI/CD generators"
      />
      <div className="mx-auto max-w-3xl w-full p-4 space-y-4">
      <section className="rounded border border-border/80 bg-card/40 p-4">
        <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Appearance</h2>
        <p className="mt-1 text-xs text-muted-foreground">Atlas themes are pure CSS-variable sets; switching themes updates design tokens instantly.</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {THEMES.map((t) => (
            <button key={t.id} onClick={() => setTheme(t.id)} className={`rounded border px-3 py-1.5 font-mono text-xs transition-colors ${theme === t.id ? "border-primary bg-primary/10 font-bold text-primary" : "border-border hover:bg-muted text-muted-foreground hover:text-foreground"}`}>{t.label}</button>
          ))}
          <button onClick={() => setTheme("system")} className={`rounded border px-3 py-1.5 font-mono text-xs transition-colors ${theme === "system" ? "border-primary bg-primary/10 font-bold text-primary" : "border-border hover:bg-muted text-muted-foreground hover:text-foreground"}`}>System</button>
        </div>
      </section>

      <ProxyTlsPanel />
      <CicdPanel />

      <section className="rounded border border-border/80 bg-card/40 p-4">
        <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Workspace</h2>
        <p className="mt-1 font-mono text-xs text-foreground">{workspace?.name ?? "No workspace"} <span className="text-muted-foreground">({workspace?.path ?? "—"})</span></p>
        <button onClick={() => void openFolder()} className="mt-2 font-mono text-xs text-primary hover:underline">Switch folder →</button>
      </section>

      <section className="rounded border border-border/80 bg-card/40 p-4">
        <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Storage & History Retention</h2>
        <p className="mt-1 font-mono text-xs text-muted-foreground">{pool.length} recent entries recorded in local SQLite database.</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {RETENTION_OPTIONS.map((o) => (
            <button
              key={o.value}
              onClick={() => saveRetention(o.value)}
              className={`rounded border px-3 py-1.5 font-mono text-xs transition-colors ${retention === o.value ? "border-primary bg-primary/10 font-bold text-primary" : "border-border hover:bg-muted text-muted-foreground hover:text-foreground"}`}
            >
              {o.label}
            </button>
          ))}
        </div>
        <p className="mt-2 text-[11px] text-muted-foreground">Stored locally in <code className="font-mono">.reqly/history.db</code>. Older records are automatically cleaned up on launch.</p>
      </section>

      <section className="rounded border border-border/80 bg-card/40 p-4">
        <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">About & Diagnostics</h2>
        <p className="mt-1 text-xs text-muted-foreground">Reqly — Local-first, Git-native API client. Zero telemetry.</p>
        <p className="mt-1 font-mono text-xs text-muted-foreground">Version <span className="font-bold text-foreground">{APP_VERSION}</span></p>
        <div className="mt-3 rounded border border-border bg-background/50 p-3">
          <p className="font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Configured Keyboard Shortcuts</p>
          <ul className="mt-2 grid gap-1.5 font-mono text-xs">
            {SHORTCUTS.map(([k, d]) => (
              <li key={k} className="flex items-center justify-between border-b border-border/40 pb-1">
                <span className="rounded bg-muted px-1.5 py-0.5 font-bold text-primary">{k}</span>
                <span className="text-muted-foreground">{d}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>
      </div>
    </div>
  );
}

