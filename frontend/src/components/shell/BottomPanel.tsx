import { useBottomPanelStore } from "#stores/useBottomPanelStore";
import { BOTTOM_PANELS, type BottomPanelId } from "#lib/bottomPanel";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { cn } from "#lib/utils";

const LABELS: Record<BottomPanelId, string> = {
  console: "Console",
  network: "Network",
  tests: "Tests",
  variables: "Variables",
  cookies: "Cookies",
};

function PanelContent({ active }: { active: BottomPanelId }) {
  const pool = useHistoryStore((s) => s.pool);
  const envs = useWorkspaceStore((s) => s.environments);
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
    return (
      <div className="p-3 text-xs">
        {envs.length === 0 ? <p className="text-muted-foreground">No environments — effective variables will appear here.</p> : <p className="text-muted-foreground">{envs.length} environments — select one to inspect resolved variables.</p>}
      </div>
    );
  }
  if (active === "tests") {
    return <p className="p-3 text-xs text-muted-foreground">No test runs yet — run a collection to see results here.</p>;
  }
  return <p className="p-3 text-xs text-muted-foreground">No cookies in this workspace.</p>;
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
