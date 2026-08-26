import { useThemeStore } from "#stores";
import { THEMES } from "#lib/themes";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";

const SHORTCUTS: [string, string][] = [
  ["⌘K", "Command palette"],
  ["⌘B", "Toggle sidebar"],
  ["⌘J", "Toggle bottom panel"],
  ["⌘W", "Close tab"],
  ["⌘1–8", "Jump to tool (Home, Requests, Envs, History, then API tools in rail order)"],
  ["⌘⏎", "Send active request"],
];

export function SettingsView() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const workspace = useWorkspaceStore((s) => s.currentWorkspace);
  const pool = useHistoryStore((s) => s.pool);
  const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
  return (
    <div className="mx-auto max-w-3xl p-6 space-y-6">
      <h1 className="text-lg font-semibold tracking-tight">Settings</h1>
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
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">Workspace</h2>
        <p className="mt-1 text-xs text-muted-foreground">{workspace?.name ?? "No workspace"} — {workspace?.path ?? "—"}</p>
        <button onClick={() => void openFolder()} className="mt-2 text-xs underline">Switch folder</button>
      </section>
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">History</h2>
        <p className="mt-1 text-xs text-muted-foreground">{pool.length} recent entries (capped at 500)</p>
      </section>
      <section className="rounded-lg border border-border bg-card p-4">
        <h2 className="text-sm font-semibold">About</h2>
        <p className="mt-1 text-xs text-muted-foreground">Reqly — local-first, zero telemetry.</p>
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
  );
}

