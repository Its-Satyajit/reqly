import { useState } from "react";
import {
	FileDown,
	FilePlus,
	FolderSearch,
	RefreshCw,
	SquareArrowOutDownLeft,
} from "lucide-react";
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
import { AuthPanel } from "../features/auth-panel/AuthPanel";
import { ExportDialog } from "../features/export-dialog/ExportDialog";
import { ImportDialog } from "../features/import-dialog/ImportDialog";
import { useRealtimeStore } from "../stores/useRealtimeStore";
import { useTestStore } from "../stores/useTestStore";
import { useExportStore } from "../stores/useExportStore";
import { useImportStore } from "../stores/useImportStore";
import { useWorkspaceStore, type WorkspaceView } from "../stores/useWorkspaceStore";
import { useWorkspaceBootstrapStore } from "../stores/useWorkspaceBootstrap";
import { CollectionTree } from "./CollectionTree";
import { GitPanel } from "./GitPanel";

export function WorkspaceSidebar() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const hasUnsavedEnvChanges = useWorkspaceStore((s) => s.hasUnsavedEnvChanges);
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const switchWorkspace = useWorkspaceBootstrapStore((s) => s.openFolder);
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
				{navItem("mocks", "Mocks")}
				{navItem("diff", "Diff")}
				{navItem("jwt", "JWT")}
				{navItem("graphql", "GraphQL")}
				{navItem("runners", "Runners")}
				{navItem("explorer", "Explorer")}
				{navItem("docs", "Docs")}
				{navItem("grpc", "gRPC")}
				{navItem("settings", "Settings")}
			</nav>
			<div className="border-t border-border pt-2">
				<div className="flex items-center justify-between px-2 pb-2">
					<p className="truncate text-xs font-medium uppercase tracking-wide text-muted-foreground">
						{workspaceName ? workspaceName : "Collections"}
					</p>
					<span className="flex items-center gap-0.5">
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
 				<CollectionTree />
				<GitPanel />
				<TestsSection />
				<RealtimeSection />
				<AuthPanel />
			</div>
			<ImportDialog
				onImported={() => void refreshWorkspace()}
			/>
			<ExportDialog />
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

function TestsSection() {
	const tests = useTestStore((s) => s.tests);
	const openPath = useTestStore((s) => s.openPath);
	const newTab = useTestStore((s) => s.newTab);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);

	const openTestTab = (path: string | null, title: string, fresh: boolean) => {
		const id = fresh ? `test-new-${Date.now()}` : `test-${path}`;
		if (fresh) newTab(id);
		else void openPath(id, path ?? "");
		openTab({ id, title, kind: "test", filePath: path ?? undefined });
		setActiveView("requests");
	};

	return (
		<div className="border-t border-border px-2 pt-2">
			<div className="flex items-center justify-between pb-1">
				<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
					Tests {tests.length > 0 && `(${tests.length})`}
				</p>
				<button
					type="button"
					onClick={() => openTestTab(null, "untitled.reqly-test", true)}
					title="New test file"
					aria-label="New test file"
					className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
				>
					<FilePlus className="size-3.5" aria-hidden />
				</button>
			</div>
			{tests.length === 0 ? (
				<p className="pb-2 text-[11px] text-muted-foreground">No *.reqly-test files found.</p>
			) : (
				<ul className="flex flex-col gap-0.5 pb-2">
					{tests.map((t) => (
						<li key={t.path}>
							<button
								type="button"
								className="w-full truncate rounded px-1 py-0.5 text-left text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
								title={t.path}
								onClick={() => openTestTab(t.path, t.name || t.path, false)}
							>
								{t.name || t.path}
							</button>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}

function RealtimeSection() {
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const newTab = useRealtimeStore((s) => s.newTab);

	const openRealtimeTab = (kind: "ws" | "sse") => {
		const id = `realtime-${kind}-${Date.now()}`;
		newTab(id, kind);
		openTab({ id, title: kind === "ws" ? "WebSocket" : "SSE", kind: "realtime" });
		setActiveView("requests");
	};

	return (
		<div className="border-t border-border px-2 pt-2">
			<p className="pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
				Realtime
			</p>
			<div className="flex gap-1 pb-2">
				<button
					type="button"
					onClick={() => openRealtimeTab("ws")}
					className="flex-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
					title="Open a WebSocket tab"
				>
					WS
				</button>
				<button
					type="button"
					onClick={() => openRealtimeTab("sse")}
					className="flex-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
					title="Open an SSE tab"
				>
					SSE
				</button>
			</div>
		</div>
	);
}
