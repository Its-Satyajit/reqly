import { AuthPanel } from "../features";
import { useWorkspaceStore } from "../stores";
import { cn } from "#lib/utils";

export function WorkspaceSidebar() {
  const activeView = useWorkspaceStore((s) => s.activeView);
  const setActiveView = useWorkspaceStore((s) => s.setActiveView);
  const hasUnsavedEnvChanges = useWorkspaceStore((s) => s.hasUnsavedEnvChanges);

  const navItem = (view: "requests" | "environments", label: string) => (
    <button
      onClick={() => {
        if (activeView === view) return;
        if (hasUnsavedEnvChanges && !window.confirm("Discard unsaved environment changes?")) {
          return;
        }
        setActiveView(view);
      }}
      className={cn(
        "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
        activeView === view
          ? "bg-muted font-medium text-foreground"
          : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
      )}
    >
      {label}
    </button>
  );

  return (
    <aside className="w-64 shrink-0 overflow-y-auto border-r border-border p-2">
      <nav className="flex flex-col gap-0.5 pb-2">
        {navItem("requests", "Requests")}
        {navItem("environments", "Environments")}
      </nav>
      <div className="border-t border-border pt-2">
        <p className="px-2 pb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Collections
        </p>
        <p className="px-2 text-xs text-muted-foreground">No collections yet</p>
        <AuthPanel />
      </div>
    </aside>
  )
}