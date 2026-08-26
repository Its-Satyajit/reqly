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
import { Button } from "#components/ui/button";
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
		<aside className="h-full w-full border-r border-border">
			<ScrollArea className="size-full">
				<div className="flex flex-col p-2">
			<div className="flex items-center justify-between px-2 pb-1">
				<p className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
					Collections
				</p>
				<span className="flex items-center gap-0.5">
					<Button variant="ghost" size="icon-xs" onClick={() => setCreateDialog({ kind: "collection" })} title="New collection" aria-label="New collection">
						<FolderPlus className="size-3.5" aria-hidden />
					</Button>
					<Button variant="ghost" size="icon-xs" onClick={openNewRequest} title="New request" aria-label="New request">
						<FilePlus className="size-3.5" aria-hidden />
					</Button>
					{workspaceTree && (
						<Button variant="ghost" size="icon-xs" onClick={() => void switchWorkspace()} title="Switch workspace" aria-label="Switch workspace">
							<FolderSearch className="size-3.5" aria-hidden />
						</Button>
					)}
					{workspaceTree && (
						<Button variant="ghost" size="icon-xs" onClick={() => setImportOpen(true)} title="Import from cURL, OpenAPI, HAR, Postman, Insomnia, or Bruno" aria-label="Import into workspace">
							<SquareArrowOutDownLeft className="size-3.5" aria-hidden />
						</Button>
					)}
					{workspaceTree && (
						<Button variant="ghost" size="icon-xs" onClick={() => setExportOpen(true)} title="Export as Postman, OpenAPI, HAR, or a workspace copy" aria-label="Export workspace">
							<FileDown className="size-3.5" aria-hidden />
						</Button>
					)}
					{workspaceTree && (
						<Button variant="ghost" size="icon-xs" onClick={() => void refreshWorkspace()} title="Reload the workspace tree from disk" aria-label="Reload workspace">
							<RefreshCw className="size-3.5" aria-hidden />
						</Button>
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
				<Button
					variant="outline"
					size="xs"
					className="flex-1"
					onClick={openNewRequest}
					title="Create a new request"
				>
					<FilePlus className="size-3" aria-hidden />
					Request
				</Button>
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
				</div>
			</ScrollArea>
		</aside>
	);
}
