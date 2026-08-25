import { useState } from "react";
import {
	FileDown,
	FilePlus,
	FolderPlus,
	FolderSearch,
	RefreshCw,
	SquareArrowOutDownLeft,
} from "lucide-react";
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
import { ExportDialog } from "../features/export-dialog/ExportDialog";
import { ImportDialog } from "../features/import-dialog/ImportDialog";
import { useExportStore } from "../stores/useExportStore";
import { useImportStore } from "../stores/useImportStore";
import { useWorkspaceStore, type WorkspaceView } from "../stores/useWorkspaceStore";
import { NEW_REQUEST_TAB_ID } from "../stores/useRequestStore";
import { useWorkspaceBootstrapStore } from "../stores/useWorkspaceBootstrap";
import { CollectionTree } from "./CollectionTree";
import { NewContainerDialog } from "./NewContainerDialog";

/** Runs a create call, mapping failures to a display message. Module-level
 * because React Compiler cannot handle try/catch. */
async function runCreate(
	create: () => Promise<void>,
): Promise<string | null> {
	try {
		await create();
		return null;
	} catch (err) {
		return err instanceof Error ? err.message : String(err);
	}
}

type CreateDialog =
	| { kind: "collection" }
	| { kind: "folder"; parent: string; parentName: string }
	| null;

export function WorkspaceSidebar() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const hasUnsavedEnvChanges = useWorkspaceStore((s) => s.hasUnsavedEnvChanges);
	const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const switchWorkspace = useWorkspaceBootstrapStore((s) => s.openFolder);
	const workspaceAdapter = useWorkspaceStore((s) => s.workspaceAdapter);
	const [pendingView, setPendingView] = useState<WorkspaceView | null>(null);
	const [createDialog, setCreateDialog] = useState<CreateDialog>(null);

	const requestView = (view: WorkspaceView) => {
		if (activeView === view) return;
		if (hasUnsavedEnvChanges) {
			setPendingView(view);
			return;
		}
		setActiveView(view);
	};

	const openNewRequest = () => {
		openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		requestView("requests");
	};

	const submitCreate = async (name: string): Promise<string | null> => {
		if (!createDialog) return null;
		const createError =
			createDialog.kind === "collection"
				? await runCreate(() => workspaceAdapter.createCollection(name))
				: await runCreate(() =>
						workspaceAdapter.createFolder(createDialog.parent, name),
					);
		if (createError !== null) return createError;
		await refreshWorkspace();
		return null;
	};

	return (
		<aside className="flex h-full w-full flex-col overflow-y-auto border-r border-border p-2">
			<div className="flex items-center justify-between px-2 pb-1">
				<p className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
					Collections
				</p>
				<span className="flex items-center gap-0.5">
					<button
						type="button"
						onClick={() => setCreateDialog({ kind: "collection" })}
						title="New collection"
						aria-label="New collection"
						className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
					>
						<FolderPlus className="size-3.5" aria-hidden />
					</button>
					<button
						type="button"
						onClick={openNewRequest}
						title="New request"
						aria-label="New request"
						className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
					>
						<FilePlus className="size-3.5" aria-hidden />
					</button>
					{workspaceTree && (
						<button
							type="button"
							onClick={() => void switchWorkspace()}
							title="Switch workspace"
							aria-label="Switch workspace"
							className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
						>
							<FolderSearch className="size-3.5" aria-hidden />
						</button>
					)}
					{workspaceTree && (
						<button
							type="button"
							onClick={() => setImportOpen(true)}
							title="Import from cURL, OpenAPI, HAR, Postman, Insomnia, or Bruno"
							aria-label="Import into workspace"
							className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
						>
							<SquareArrowOutDownLeft className="size-3.5" aria-hidden />
						</button>
					)}
					{workspaceTree && (
						<button
							type="button"
							onClick={() => setExportOpen(true)}
							title="Export as Postman, OpenAPI, HAR, or a workspace copy"
							aria-label="Export workspace"
							className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
						>
							<FileDown className="size-3.5" aria-hidden />
						</button>
					)}
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
				</span>
			</div>
 			<CollectionTree
				onNewCollection={() => setCreateDialog({ kind: "collection" })}
				onNewFolder={(path, name) =>
					setCreateDialog({ kind: "folder", parent: path, parentName: name })
				}
			/>
			<div className="mt-2 flex shrink-0 gap-1">
				<button
					type="button"
					onClick={openNewRequest}
					className="flex flex-1 items-center justify-center gap-1.5 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
					title="Create a new request"
				>
					<FilePlus className="size-3" aria-hidden />
					Request
				</button>
			</div>
			<ImportDialog
				onImported={() => void refreshWorkspace()}
			/>
			<ExportDialog />
			{createDialog && (
				<NewContainerDialog
				title={
					createDialog?.kind === "folder"
						? `New folder in ${createDialog.parentName}`
						: "New collection"
				}
				description={
					createDialog?.kind === "folder"
						? "Folders nest requests inside a collection. The descriptor lands on disk immediately — Git-native, no cloud."
						: "Collections are plain folders versioned with Git. The descriptor lands on disk immediately."
				}
					onCreate={submitCreate}
					onClose={() => setCreateDialog(null)}
				/>
			)}
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
