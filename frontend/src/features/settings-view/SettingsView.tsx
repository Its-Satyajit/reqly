import { useState, useEffect } from "react";
import {
  Settings,
  FolderGit2,
  ExternalLink,
  Keyboard,
  Database,
  Info,
} from "lucide-react";
import { useThemeStore, useSettingsStore } from "#stores";
import { THEMES } from "#lib/themes";
import { APP_VERSION } from "#lib/crash";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";
import { PageHeader } from "#components/PageHeader";
import { ProxyPanel, TlsSecurityPanel } from "./ProxyTlsPanel";
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

const SETTINGS_DESCRIPTIONS = {
  appearance: "Customize color theme, design tokens, and display preferences",
  workspace: "Active workspace directory, Git repository status, and folder selection",
  storage: "Local SQLite database retention policies and cache management",
  network: "Global upstream proxy server and bypass configurations",
  security: "SSL / TLS certificates, peer verification, and mTLS client credentials",
  cicd: "Automated test pipeline generators and CI workflow scripts",
  shortcuts: "Reference list of configured global keyboard shortcuts",
  about: "Version details, licenses, and zero-telemetry environment info",
} satisfies Record<import("#stores").SettingsTabId, string>;

export function SettingsView() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const workspace = useWorkspaceStore((s) => s.currentWorkspace);
  const pool = useHistoryStore((s) => s.pool);
  const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
  const activeTab = useSettingsStore((s) => s.activeTab);
  const [retention, setRetention] = useState(() => {
    if (typeof window === "undefined") return "90d";
    return localStorage.getItem("reqly:historyRetention") || "90d";
  });

  const saveRetention = (v: string) => {
    setRetention(v);
    localStorage.setItem("reqly:historyRetention", v);
  };

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-background" aria-label="Settings">
      <PageHeader
        icon={Settings}
        title="Settings"
        description={SETTINGS_DESCRIPTIONS[activeTab] || "Application preferences and configurations"}
      />

      <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4">
        <div className="mx-auto max-w-4xl w-full">
          {activeTab === "appearance" && (
            <div className="space-y-4">
              <section className="rounded-lg border border-border/80 bg-card/40 p-5 space-y-4">
                <div>
                  <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Themes & Colors</h2>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Atlas themes are pure CSS-variable sets. Switching themes updates all application tokens instantly.
                  </p>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2.5 pt-2">
                  {THEMES.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => setTheme(t.id)}
                      className={`flex flex-col items-start rounded-md border p-3 font-mono text-xs text-left transition-all ${
                        theme === t.id
                          ? "border-primary bg-primary/10 font-bold text-primary ring-1 ring-primary"
                          : "border-border hover:bg-muted/60 text-muted-foreground hover:text-foreground"
                      }`}
                    >
                      <span className="font-medium text-foreground">{t.label}</span>
                      <span className="mt-1 text-[10px] text-muted-foreground">{t.id}</span>
                    </button>
                  ))}
                  <button
                    onClick={() => setTheme("system")}
                    className={`flex flex-col items-start rounded-md border p-3 font-mono text-xs text-left transition-all ${
                      theme === "system"
                        ? "border-primary bg-primary/10 font-bold text-primary ring-1 ring-primary"
                        : "border-border hover:bg-muted/60 text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <span className="font-medium text-foreground">System</span>
                    <span className="mt-1 text-[10px] text-muted-foreground">Follow OS theme</span>
                  </button>
                </div>
              </section>
            </div>
          )}

          {activeTab === "workspace" && (
            <div className="space-y-4">
              <section className="rounded-lg border border-border/80 bg-card/40 p-5 space-y-4">
                <div className="flex items-center gap-2">
                  <FolderGit2 className="size-4 text-primary" />
                  <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Active Workspace Directory</h2>
                </div>
                <p className="text-xs text-muted-foreground">
                  Reqly operates directly on the local filesystem. All requests, collections, and environments live in plain text files in this folder.
                </p>
                <div className="rounded-md border border-border bg-background/60 p-3 space-y-1">
                  <div className="font-mono text-xs font-medium text-foreground">{workspace?.name ?? "No workspace open"}</div>
                  <div className="font-mono text-[11px] text-muted-foreground break-all">{workspace?.path ?? "—"}</div>
                </div>
                <button
                  type="button"
                  onClick={() => void openFolder()}
                  className="inline-flex items-center gap-1.5 rounded border border-border bg-muted/50 px-3 py-1.5 font-mono text-xs text-primary hover:bg-muted transition-colors"
                >
                  <ExternalLink className="size-3.5" />
                  Open Different Folder / Repository →
                </button>
              </section>
            </div>
          )}

          {activeTab === "storage" && (
            <div className="space-y-4">
              <section className="rounded-lg border border-border/80 bg-card/40 p-5 space-y-4">
                <div className="flex items-center gap-2">
                  <Database className="size-4 text-primary" />
                  <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Storage & History Retention</h2>
                </div>
                <p className="text-xs text-muted-foreground">
                  <span className="font-mono font-medium text-foreground">{pool.length}</span> recent request execution entries stored in local SQLite database (<code className="font-mono text-foreground/80">.reqly/history.db</code>).
                </p>
                <div className="flex flex-wrap gap-2 pt-1">
                  {RETENTION_OPTIONS.map((o) => (
                    <button
                      key={o.value}
                      onClick={() => saveRetention(o.value)}
                      className={`rounded border px-3.5 py-1.5 font-mono text-xs transition-colors ${
                        retention === o.value
                          ? "border-primary bg-primary/10 font-bold text-primary ring-1 ring-primary"
                          : "border-border hover:bg-muted text-muted-foreground hover:text-foreground"
                      }`}
                    >
                      {o.label}
                    </button>
                  ))}
                </div>
                <p className="text-[11px] text-muted-foreground">
                  Older records beyond the retention threshold are automatically pruned from SQLite on launch.
                </p>
              </section>
            </div>
          )}

          {activeTab === "network" && <ProxyPanel />}

          {activeTab === "security" && <TlsSecurityPanel />}

          {activeTab === "cicd" && <CicdPanel />}

          {activeTab === "shortcuts" && (
            <div className="space-y-4">
              <section className="rounded-lg border border-border/80 bg-card/40 p-5 space-y-4">
                <div className="flex items-center gap-2">
                  <Keyboard className="size-4 text-primary" />
                  <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">Global Keyboard Shortcuts</h2>
                </div>
                <p className="text-xs text-muted-foreground">
                  Keyboard combinations for high-speed navigation and execution across Reqly.
                </p>
                <div className="rounded-lg border border-border bg-background/50 p-4 mt-2">
                  <ul className="grid gap-2 font-mono text-xs">
                    {SHORTCUTS.map(([k, d]) => (
                      <li key={k} className="flex items-center justify-between border-b border-border/40 pb-1.5 last:border-0 last:pb-0">
                        <span className="rounded bg-muted px-2 py-0.5 font-bold text-primary border border-border/60">{k}</span>
                        <span className="text-muted-foreground">{d}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </section>
            </div>
          )}

          {activeTab === "about" && (
            <div className="space-y-4">
              <section className="rounded-lg border border-border/80 bg-card/40 p-5 space-y-4">
                <div className="flex items-center gap-2">
                  <Info className="size-4 text-primary" />
                  <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-muted-foreground">About Reqly</h2>
                </div>
                <p className="text-xs text-muted-foreground">
                  Reqly is a local-first, Git-native API development and testing environment. All environments, collections, files, and secrets remain entirely local on your machine.
                </p>
                <div className="rounded-md border border-border bg-background/60 p-4 space-y-2">
                  <div className="flex items-center justify-between font-mono text-xs">
                    <span className="text-muted-foreground">App Version</span>
                    <span className="font-bold text-primary">{APP_VERSION}</span>
                  </div>
                  <div className="flex items-center justify-between font-mono text-xs">
                    <span className="text-muted-foreground">Architecture</span>
                    <span className="text-foreground">Go Core + Wails v3 + React 19</span>
                  </div>
                  <div className="flex items-center justify-between font-mono text-xs">
                    <span className="text-muted-foreground">Telemetry</span>
                    <span className="text-foreground font-semibold text-emerald-500">Zero Telemetry (Disabled)</span>
                  </div>
                  <div className="flex items-center justify-between font-mono text-xs">
                    <span className="text-muted-foreground">License</span>
                    <span className="text-foreground">Apache-2.0</span>
                  </div>
                </div>
              </section>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}



