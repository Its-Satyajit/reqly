import { useBottomPanelStore } from "#stores/useBottomPanelStore";
import { BOTTOM_PANELS, type BottomPanelId } from "#lib/bottomPanel";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useRequestStore } from "#stores/useRequestStore";
import { useCollectionRunStore } from "#stores/useCollectionRunStore";
import { parseSetCookies, type ResponseCookie } from "#lib/response";
import type { ResolvedVariable } from "#lib/collections";
import { cn } from "#lib/utils";

const LABELS: Record<BottomPanelId, string> = {
  console: "Console",
  network: "Network",
  tests: "Tests",
  variables: "Variables",
  cookies: "Cookies",
  devtools: "DevTools",
};

function PanelContent({ active }: { active: BottomPanelId }) {
  const pool = useHistoryStore((s) => s.pool);
  const envs = useWorkspaceStore((s) => s.environments);
  const activeEnvId = useWorkspaceStore((s) => s.activeEnvironmentId);
  const activeTabId = useWorkspaceStore((s) => s.activeTabId);
  const tabMeta = useRequestStore((s) => (activeTabId ? s.meta[activeTabId] : undefined));
  const activeTabResponse = useRequestStore((s) => (activeTabId ? s.responses[activeTabId] : undefined));
  const runSteps = useCollectionRunStore((s) => s.steps);
  const runReport = useCollectionRunStore((s) => s.report);

  if (active === "console") {
    return (
      <div className="p-3 font-mono text-xs leading-relaxed">
        {pool.length === 0 ? (
          <p className="text-muted-foreground">No logs yet — send a request to see the trace.</p>
        ) : (
          <ul className="space-y-1">
            {pool.slice(0, 8).map((e) => (
              <li key={e.id} className="flex gap-2">
                <span className="text-muted-foreground">{new Date(e.createdAt).toLocaleTimeString()}</span>
                <span className="text-status-info">INFO</span>
                <span>Sending {e.method} {e.url}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    );
  }
  if (active === "network") {
    return (
      <div className="p-2">
        {pool.length === 0 ? (
          <p className="px-1 text-xs text-muted-foreground">No network activity in this workspace.</p>
        ) : (
          <table className="w-full text-xs">
            <thead className="text-left text-muted-foreground">
              <tr><th className="px-2 py-1 font-medium">Time</th><th className="px-2 py-1 font-medium">Method</th><th className="px-2 py-1 font-medium">URL</th><th className="px-2 py-1 font-medium">Status</th></tr>
            </thead>
            <tbody className="font-mono">
              {pool.slice(0, 10).map((e) => (
                <tr key={e.id} className="border-t border-border/50">
                  <td className="px-2 py-1 text-muted-foreground">{new Date(e.createdAt).toLocaleTimeString()}</td>
                  <td className="px-2 py-1">{e.method}</td>
                  <td className="px-2 py-1 truncate max-w-[320px]">{e.url}</td>
                  <td className="px-2 py-1">{e.status ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    );
  }
  if (active === "variables") {
    const activeEnv = envs.find((e) => e.id === activeEnvId);
    const envVars = activeEnv ? Object.entries(activeEnv.variables) : [];
    const resolvedVars = tabMeta?.variables ?? [];

    if (envVars.length === 0 && resolvedVars.length === 0) {
      return (
        <div className="p-3 text-xs text-muted-foreground">
          {activeEnvId ? `Environment "${activeEnv?.name}" has no variables.` : "No active environment selected."}
        </div>
      );
    }

    return (
      <div className="p-3 text-xs space-y-3">
        {activeEnv && (
          <div>
            <p className="font-semibold text-muted-foreground mb-1.5 uppercase tracking-wider text-[10px]">
              Environment: {activeEnv.name} ({envVars.length} variables)
            </p>
            <div className="rounded border border-border bg-background overflow-hidden font-mono">
              <table className="w-full text-xs">
                <tbody>
                  {envVars.map(([k, v]) => (
                    <tr key={k} className="border-b border-border/50 last:border-b-0">
                      <td className="px-2.5 py-1 text-primary w-1/3 truncate">{k}</td>
                      <td className="px-2.5 py-1 text-muted-foreground truncate">{activeEnv.secrets?.includes(k) ? "••••••••" : v}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {resolvedVars.length > 0 && (
          <div>
            <p className="font-semibold text-muted-foreground mb-1.5 uppercase tracking-wider text-[10px]">
              Tab Resolved Chain ({resolvedVars.length})
            </p>
            <div className="rounded border border-border bg-background overflow-hidden font-mono">
              <table className="w-full text-xs">
                <tbody>
                  {resolvedVars.map((r: ResolvedVariable, i: number) => (
                    <tr key={i} className="border-b border-border/50 last:border-b-0">
                      <td className="px-2.5 py-1 text-primary w-1/3 truncate">{r.name}</td>
                      <td className="px-2.5 py-1 text-muted-foreground truncate">{r.value}</td>
                      <td className="px-2.5 py-1 text-[10px] text-muted-foreground/60 w-20 text-right">{r.scope}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    );
  }
  if (active === "tests") {
    if (runSteps.length === 0 && !runReport) {
      return <p className="p-3 text-xs text-muted-foreground">No test runs yet — run a collection to see results here.</p>;
    }
    return (
      <div className="p-3 text-xs space-y-2">
        {runReport && (
          <div className="flex items-center gap-3 text-xs font-mono">
            <span className="text-status-ok">{runReport.passed} passed</span>
            <span className="text-destructive">{runReport.failed} failed</span>
            <span className="text-muted-foreground">{runReport.durationMs}ms</span>
          </div>
        )}
        <ul className="space-y-1 font-mono text-xs">
          {runSteps.map((st, i) => (
            <li key={i} className="flex items-center gap-2 border-b border-border/40 py-1 last:border-0">
              <span className={st.passed ? "text-status-ok" : "text-destructive font-semibold"}>
                {st.passed ? "✓ PASS" : "✕ FAIL"}
              </span>
              <span className="truncate">{st.name}</span>
              {st.response?.durationMs != null && (
                <span className="ml-auto text-[10px] text-muted-foreground">{st.response.durationMs}ms</span>
              )}
            </li>
          ))}
        </ul>
      </div>
    );
  }
  if (active === "cookies") {
    const rawHeaders = activeTabResponse?.response?.headers;
    const cookies: ResponseCookie[] = rawHeaders ? parseSetCookies(rawHeaders) : [];
    if (cookies.length === 0) {
      return <p className="p-3 text-xs text-muted-foreground">No cookies in active response.</p>;
    }
    return (
      <div className="p-2">
        <table className="w-full text-xs">
          <thead className="text-left text-muted-foreground">
            <tr><th className="px-2 py-1 font-medium">Name</th><th className="px-2 py-1 font-medium">Value</th><th className="px-2 py-1 font-medium">Domain</th><th className="px-2 py-1 font-medium">Path</th></tr>
          </thead>
          <tbody className="font-mono">
            {cookies.map((c: ResponseCookie, idx: number) => (
              <tr key={idx} className="border-t border-border/50">
                <td className="px-2 py-1 text-primary">{c.name}</td>
                <td className="px-2 py-1 truncate max-w-[200px]">{c.value}</td>
                <td className="px-2 py-1 text-muted-foreground">{c.domain ?? "—"}</td>
                <td className="px-2 py-1 text-muted-foreground">{c.path ?? "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }
  // DevTools — read-only request/auth/variables/console/network per spec m54
  return (
    <div className="grid grid-cols-2 gap-2 p-2 text-xs">
      <div className="rounded border border-border p-2"><p className="font-medium">Request</p><p className="text-muted-foreground">Open a request tab to inspect.</p></div>
      <div className="rounded border border-border p-2"><p className="font-medium">Auth</p><p className="text-muted-foreground">Inherited auth applied silently.</p></div>
      <div className="rounded border border-border p-2"><p className="font-medium">Variables</p><p className="text-muted-foreground">Effective variables per tab.</p></div>
      <div className="rounded border border-border p-2"><p className="font-medium">Console</p><p className="text-muted-foreground">Goja console per run.</p></div>
      <div className="col-span-2 rounded border border-border p-2"><p className="font-medium">Network</p><p className="text-muted-foreground">Timings waterfall via ResponseViewer Timeline.</p></div>
    </div>
  );
}

export function BottomPanel() {
  const active = useBottomPanelStore((s) => s.active);
  const collapsed = useBottomPanelStore((s) => s.collapsed);
  const toggle = useBottomPanelStore((s) => s.toggle);
  return (
    <div className="flex h-full flex-col border-t border-border bg-card/30">
      <div className="flex items-center gap-1 border-b border-border px-2 py-1">
        {BOTTOM_PANELS.map((id) => (
          <button
            key={id}
            type="button"
            onClick={() => toggle(id)}
            aria-pressed={active === id && !collapsed}
            className={cn(
              "rounded-md px-2.5 py-1 text-xs font-medium transition-colors",
              active === id && !collapsed ? "bg-primary/12 text-primary" : "text-muted-foreground hover:bg-muted hover:text-foreground",
            )}
          >
            {LABELS[id]}
          </button>
        ))}
        <span className="ml-auto text-[11px] text-muted-foreground">⌘J to toggle</span>
      </div>
      {!collapsed && active ? <div className="min-h-0 flex-1 overflow-auto"><PanelContent active={active} /></div> : null}
    </div>
  );
}
