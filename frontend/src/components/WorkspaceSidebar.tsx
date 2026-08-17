import { AuthPanel } from "../features";

export function WorkspaceSidebar() {
  return (
    <aside className="w-64 shrink-0 overflow-y-auto border-r border-border p-2">
      <p className="px-2 pb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
        Collections
      </p>
      <p className="px-2 text-xs text-muted-foreground">No collections yet</p>
      <AuthPanel />
    </aside>
  )
}