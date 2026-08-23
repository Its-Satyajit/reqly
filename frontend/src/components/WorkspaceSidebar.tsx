import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { cn } from "#lib/utils";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#components/ui/alert-dialog";
import { AuthPanel } from "../features";
import { useWorkspaceStore, type WorkspaceView } from "../stores";
import { CollectionTree } from "./CollectionTree";

export function WorkspaceSidebar() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const hasUnsavedEnvChanges = useWorkspaceStore((s) => s.hasUnsavedEnvChanges);
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const [pendingView, setPendingView] = useState<WorkspaceView | null>(null);

	const requestView = (view: WorkspaceView) => {
		if (activeView === view) return;
		if (hasUnsavedEnvChanges) {
			setPendingView(view);
			return;
		}
		setActiveView(view);
	};

	const navItem = (view: WorkspaceView, label: string) => (
		<button
			type="button"
			aria-current={activeView === view ? "page" : undefined}
			onClick={() => requestView(view)}
			className={cn(
				"flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
				activeView === view
					? "bg-muted font-medium text-foreground"
					: "text-muted-foreground hover:bg-muted/50 hover:text-foreground",
			)}
		>
			{label}
		</button>
	);

	return (
		<aside className="flex h-full w-full flex-col overflow-y-auto border-r border-border p-2">
			<nav aria-label="Workspace views" className="flex flex-col gap-0.5 pb-2">
				{navItem("requests", "Requests")}
				{navItem("environments", "Environments")}
				{navItem("history", "History")}
			</nav>
			<div className="border-t border-border pt-2">
				<div className="flex items-center justify-between px-2 pb-2">
					<p className="truncate text-xs font-medium uppercase tracking-wide text-muted-foreground">
						{workspaceName ? workspaceName : "Collections"}
					</p>
					{workspaceTree && (
						<button
							type="button"
							onClick={() => void refreshWorkspace()}
							title="Reload the workspace tree from disk"
							aria-label="Reload workspace"
							className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
						>
							<RefreshCw className="size-3.5" aria-hidden />
						</button>
					)}
				</div>
				<CollectionTree />
				<AuthPanel />
			</div>
			<AlertDialog
				open={pendingView != null}
				onOpenChange={(open) => {
					if (!open) setPendingView(null);
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Discard unsaved environment changes?</AlertDialogTitle>
						<AlertDialogDescription>
							Switching views discards edits that were never saved.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Keep editing</AlertDialogCancel>
						<AlertDialogAction
							onClick={() => {
								if (pendingView) setActiveView(pendingView);
								setPendingView(null);
							}}
						>
							Discard changes
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</aside>
	);
}
