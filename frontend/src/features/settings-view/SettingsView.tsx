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
  ["⌘K", "Command palette"],
  ["⌘B", "Toggle sidebar"],
  ["⌘J", "Toggle bottom panel"],
  ["⌘W", "Close tab"],
  ["⌘1–8", "Jump to tool (Home, Requests, Envs, History, then API tools in rail order)"],
  ["⌘⏎", "Send active request"],
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
    <div className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="Settings">
      <PageHeader
        icon={Settings}
        title="Settings"
        description="Preferences, workspace configurations, proxy/TLS rules, and CI/CD generators"
      />
      <div className="mx-auto max-w-3xl w-full p-6 space-y-6">
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">Appearance</h2>
        <p className="mt-1 text-xs text-muted-foreground">Atlas themes are pure CSS-variable sets; adding a theme is one stylesheet.</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {THEMES.map((t) => (
            <button key={t.id} onClick={() => setTheme(t.id)} className={`rounded-md border px-3 py-1.5 text-xs ${theme === t.id ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted"}`}>{t.label}</button>
          ))}
          <button onClick={() => setTheme("system")} className={`rounded-md border px-3 py-1.5 text-xs ${theme === "system" ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted"}`}>System</button>
        </div>
      </section>

      <ProxyTlsPanel />
      <CicdPanel />
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">Workspace</h2>
        <p className="mt-1 text-xs text-muted-foreground">{workspace?.name ?? "No workspace"} — {workspace?.path ?? "—"}</p>
        <button onClick={() => void openFolder()} className="mt-2 text-xs underline">Switch folder</button>
      </section>
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">Storage — History Retention</h2>
        <p className="mt-1 text-xs text-muted-foreground">{pool.length} recent entries (capped at 500) — retention cleanup runs on app start.</p>
        <div className="mt-3 flex gap-2">
          {RETENTION_OPTIONS.map((o) => (
            <button
              key={o.value}
              onClick={() => saveRetention(o.value)}
              className={`rounded-md border px-3 py-1.5 text-xs ${retention === o.value ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted"}`}
            >
              {o.label}
            </button>
          ))}
        </div>
        <p className="mt-2 text-xs text-muted-foreground">Setting stored locally; history entries older than retention are cleaned up on next launch.</p>
      </section>
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">About</h2>
        <p className="mt-1 text-xs text-muted-foreground">Reqly — local-first, zero telemetry.</p>
        <p className="mt-1 text-xs text-muted-foreground">Version <span className="font-mono font-medium text-foreground">{APP_VERSION}</span> — report bugs with this version.</p>
        <div className="mt-3 rounded-md border border-border bg-muted/30 p-3">
          <p className="text-xs font-medium">Keyboard shortcuts</p>
          <ul className="mt-2 grid gap-1">
            {SHORTCUTS.map(([k, d]) => (
              <li key={k} className="flex justify-between text-xs"><span className="font-mono text-muted-foreground">{k}</span><span>{d}</span></li>
            ))}
          </ul>
        </div>
      </section>
      </div>
    </div>
  );
}

