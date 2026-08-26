import { useThemeStore } from "#stores";
import { THEMES } from "#lib/themes";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";

export function SettingsView() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const workspace = useWorkspaceStore((s) => s.currentWorkspace);
  const pool = useHistoryStore((s) => s.pool);
  const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
  return (
    <div className="p-6 space-y-8 max-w-2xl">
      <section>
        <h2 className="text-sm font-semibold">Appearance</h2>
        <div className="mt-2 flex gap-2">
          {THEMES.map((t) => (
            <button key={t.id} onClick={() => setTheme(t.id)} className={theme === t.id ? "font-bold" : ""}>{t.label}</button>
          ))}
          <button onClick={() => setTheme("system")} className={theme === "system" ? "font-bold" : ""}>System</button>
        </div>
      </section>
      <section>
        <h2 className="text-sm font-semibold">Workspace</h2>
        <p className="text-xs text-muted-foreground">{workspace?.name ?? "No workspace"} — {workspace?.path ?? ""}</p>
        <button onClick={() => void openFolder()} className="text-xs underline">Switch folder</button>
      </section>
      <section>
        <h2 className="text-sm font-semibold">History</h2>
        <p className="text-xs text-muted-foreground">{pool.length} recent entries</p>
      </section>
      <section>
        <h2 className="text-sm font-semibold">About</h2>
        <p className="text-xs text-muted-foreground">Reqly — keyboard: ⌘K palette, ⌘B sidebar, ⌘W close, ⌘1–8 jumps, ⌘⏎ send</p>
      </section>
    </div>
  );
}
