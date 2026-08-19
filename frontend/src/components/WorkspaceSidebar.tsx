import { AuthPanel } from "../features";
import { useWorkspaceStore } from "../stores";
import { CollectionTree } from "./CollectionTree";
import { cn } from "#lib/utils";

export function WorkspaceSidebar() {
  const activeView = useWorkspaceStore((s) => s.activeView);
  const setActiveView = useWorkspaceStore((s) => s.setActiveView);
  const hasUnsavedEnvChanges = useWorkspaceStore((s) => s.hasUnsavedEnvChanges);
  const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
  const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
  const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);

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
        <div className="flex items-center justify-between px-2 pb-2">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {workspaceName ? workspaceName : "Collections"}
          </p>
          {workspaceTree && (
            <button
              onClick={() => void refreshWorkspace()}
              title="Reload the workspace tree from disk"
              className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
            >
              <svg
                className="size-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <path d="M21 12a9 9 0 1 1-2.64-6.36" />
                <path d="M21 3v6h-6" />
              </svg>
            </button>
          )}
        </div>
        <CollectionTree />
        <AuthPanel />
      </div>
    </aside>
  )
}