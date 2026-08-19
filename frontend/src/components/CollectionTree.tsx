import type { WorkspaceFolder } from "#lib/collections";
import type { WorkspaceRequest } from "#lib/collections";
import { cn } from "#lib/utils";
import { useWorkspaceStore } from "#stores";

interface Props {
  folders: WorkspaceFolder[];
  requests: WorkspaceRequest[];
}

function RequestRow({ request }: { request: WorkspaceRequest }) {
  return (
    <button
      onClick={() => {
        // Opening a request into a tab lands in T3; the row is inert for now.
      }}
      className="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-xs text-muted-foreground hover:bg-muted/50 hover:text-foreground"
    >
      <span className="size-3 shrink-0 text-muted-foreground/60" aria-hidden>
        ▸
      </span>
      <span className="truncate">{request.name}</span>
    </button>
  );
}

function FolderBranch({ folder, depth }: { folder: WorkspaceFolder; depth: number }) {
  const expanded = useWorkspaceStore((s) => s.expanded[folder.path] ?? false);
  const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

  return (
    <div>
      <button
        onClick={() => toggleExpanded(folder.path)}
        className="flex w-full items-center gap-1 rounded-md px-2 py-1 text-left text-xs text-muted-foreground hover:bg-muted/50 hover:text-foreground"
        style={{ paddingLeft: `${0.5 + depth * 0.75}rem` }}
      >
        <span
          className={cn("text-muted-foreground/60 transition-transform", expanded && "rotate-90")}
          aria-hidden
        >
          ▸
        </span>
        <span className="truncate">{folder.name}</span>
      </button>
      {expanded && (
        <div className="ml-1 border-l border-border pl-1">
          <CollectionBranch folders={folder.folders} requests={folder.requests} depth={depth + 1} />
        </div>
      )}
    </div>
  );
}

function CollectionBranch({ folders, requests, depth }: Props & { depth: number }) {
  return (
    <div className="flex flex-col gap-0.5">
      {requests.map((request) => (
        <div key={request.path} style={{ paddingLeft: `${depth * 0.75}rem` }}>
          <RequestRow request={request} />
        </div>
      ))}
      {folders.map((folder) => (
        <FolderBranch key={folder.path} folder={folder} depth={depth} />
      ))}
    </div>
  );
}

export function CollectionTree() {
  const tree = useWorkspaceStore((s) => s.workspaceTree);
  const workspaceError = useWorkspaceStore((s) => s.workspaceError);
  const expanded = useWorkspaceStore((s) => s.expanded);
  const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

  if (workspaceError) {
    return <p className="px-2 text-xs text-destructive">{workspaceError}</p>;
  }
  if (!tree || tree.collections.length === 0) {
    return (
      <p className="px-2 text-xs text-muted-foreground">
        {tree?.name ? "No collections yet" : "Open a reqly workspace in the desktop app to browse collections."}
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-0.5">
      {tree.collections.map((collection) => {
        const isOpen = expanded[collection.path] ?? false;
        return (
          <div key={collection.path}>
            <button
              onClick={() => toggleExpanded(collection.path)}
              className="flex w-full items-center gap-1 rounded-md px-2 py-1 text-left text-xs font-medium text-foreground hover:bg-muted/50"
            >
              <span
                className={cn("text-muted-foreground/60 transition-transform", isOpen && "rotate-90")}
                aria-hidden
              >
                ▸
              </span>
              <span className="truncate">{collection.name}</span>
            </button>
            {isOpen && (
              <div className="ml-1 border-l border-border pl-1">
                <CollectionBranch folders={collection.folders} requests={collection.requests} depth={1} />
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}