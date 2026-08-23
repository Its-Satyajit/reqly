import { ChevronRight, Play } from "lucide-react";
import type { WorkspaceFolder, WorkspaceRequest } from "#lib/collections";
import { cn } from "#lib/utils";
import { RUN_TAB_ID, useCollectionRunStore, useWorkspaceStore } from "#stores";

const TREE_KEYS = new Set([
	"ArrowDown",
	"ArrowUp",
	"ArrowRight",
	"ArrowLeft",
	"Home",
	"End",
]);

/** Roving keyboard navigation over every visible tree row: arrows move
 * focus, Right expands the focused folder, Left collapses it. */
function treeKeyDown(event: React.KeyboardEvent<HTMLElement>): void {
	if (!TREE_KEYS.has(event.key)) return;
	const rows = Array.from(
		event.currentTarget.querySelectorAll<HTMLElement>("[data-tree-row]"),
	);
	// SAFETY: activeElement is matched against this same query's rows; foreign
	// nodes yield -1 and every branch below guards for that.
	const current = document.activeElement as HTMLElement | null;
	const index = current ? rows.indexOf(current) : -1;
	if (event.key === "ArrowDown" || event.key === "ArrowUp") {
		event.preventDefault();
		const delta = event.key === "ArrowDown" ? 1 : -1;
		const next = index === -1 ? 0 : Math.min(rows.length - 1, Math.max(0, index + delta));
		rows[next]?.focus();
		return;
	}
	if (event.key === "Home") {
		event.preventDefault();
		rows[0]?.focus();
		return;
	}
	if (event.key === "End") {
		event.preventDefault();
		rows[rows.length - 1]?.focus();
		return;
	}
	if (
		(event.key === "ArrowRight" || event.key === "ArrowLeft") &&
		current?.getAttribute("aria-expanded") != null &&
		index !== -1
	) {
		const isOpen = current.getAttribute("aria-expanded") === "true";
		if ((event.key === "ArrowRight" && !isOpen) || (event.key === "ArrowLeft" && isOpen)) {
			event.preventDefault();
			current.click();
		}
	}
}

/** RunControl is the play button on collection/folder rows: it opens the Run
 * View tab and starts a collection run, disabled while a run is in flight. */
function RunControl({ path, name }: { path: string; name: string }) {
	const running = useCollectionRunStore((s) => s.running);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const startRun = useCollectionRunStore((s) => s.startRun);

	const run = () => {
		openTab({ id: RUN_TAB_ID, title: name, kind: "run" });
		const { failFast } = useCollectionRunStore.getState();
		void startRun(path, activeEnvironmentId, failFast);
	};

	return (
		<button
			type="button"
			onClick={run}
			disabled={running}
			title={running ? "A run is already in progress" : `Run ${name}`}
			aria-label={`Run ${name}`}
			className="shrink-0 rounded p-1 text-muted-foreground/60 hover:bg-muted/50 hover:text-status-ok disabled:cursor-not-allowed disabled:opacity-40"
		>
			<Play className="size-3 fill-current" aria-hidden />
		</button>
	);
}

function RequestRow({ request }: { request: WorkspaceRequest }) {
	const openRequest = useWorkspaceStore((s) => s.openRequest);
	return (
		<div style={{ paddingLeft: "1.5rem" }}>
			<button
				type="button"
				data-tree-row
				onClick={() => void openRequest(request.path)}
				title={`Open ${request.name}`}
				className="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-xs text-muted-foreground hover:bg-muted/50 hover:text-foreground"
			>
				<span className="size-3 shrink-0" aria-hidden />
				<span className="truncate">{request.name}</span>
			</button>
		</div>
	);
}

interface BranchProps {
	folders: WorkspaceFolder[];
	requests: WorkspaceRequest[];
}

function CollectionBranch({ folders, requests }: BranchProps) {
	return (
		<div className="flex flex-col gap-0.5">
			{requests.map((request) => (
				<RequestRow key={request.path} request={request} />
			))}
			{folders.map((folder) => (
				<FolderBranch key={folder.path} folder={folder} />
			))}
		</div>
	);
}

function FolderBranch({ folder }: { folder: WorkspaceFolder }) {
	const expanded = useWorkspaceStore((s) => s.expanded[folder.path] ?? false);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

	return (
		<div>
			<div className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50">
				<button
					type="button"
					data-tree-row
					aria-expanded={expanded}
					onClick={() => toggleExpanded(folder.path)}
					className="flex min-w-0 flex-1 items-center gap-1 text-left text-xs text-muted-foreground"
				>
					<ChevronRight
						className={cn(
							"size-3 shrink-0 transition-transform",
							expanded && "rotate-90",
						)}
						aria-hidden
					/>
					<span className="truncate">{folder.name}</span>
				</button>
				<RunControl path={folder.path} name={folder.name} />
			</div>
			{expanded && (
				<div className="ml-1 border-l border-border pl-1">
					<CollectionBranch folders={folder.folders} requests={folder.requests} />
				</div>
			)}
		</div>
	);
}

export function CollectionTree() {
	const tree = useWorkspaceStore((s) => s.workspaceTree);
	const workspaceError = useWorkspaceStore((s) => s.workspaceError);
	const openError = useWorkspaceStore((s) => s.openError);
	const expanded = useWorkspaceStore((s) => s.expanded);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

	if (workspaceError) {
		return <p className="px-2 text-xs text-destructive">{workspaceError}</p>;
	}
	if (!tree || tree.collections.length === 0) {
		return (
			<p className="px-2 pb-4 text-xs leading-relaxed text-muted-foreground">
				{tree?.name
					? "No collections yet — create collections/<name>/reqly.yaml to see them here."
					: "Open a reqly workspace in the desktop app to browse collections."}
			</p>
		);
	}

	return (
		<div
			role="tree"
			aria-label="Collections"
			onKeyDown={treeKeyDown}
			className="flex flex-col gap-0.5"
		>
			{openError && (
				<p role="alert" className="px-2 text-xs text-destructive">
					{openError}
				</p>
			)}
			{tree.collections.map((collection) => {
				const isOpen = expanded[collection.path] ?? false;
				return (
					<div key={collection.path}>
						<div className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50">
							<button
								type="button"
								data-tree-row
								aria-expanded={isOpen}
								onClick={() => toggleExpanded(collection.path)}
								className="flex min-w-0 flex-1 items-center gap-1 text-left text-xs font-medium text-foreground"
							>
								<ChevronRight
									className={cn(
										"size-3 shrink-0 transition-transform",
										isOpen && "rotate-90",
									)}
									aria-hidden
								/>
								<span className="truncate">{collection.name}</span>
							</button>
							<RunControl path={collection.path} name={collection.name} />
						</div>
						{isOpen && (
							<div className="ml-1 border-l border-border pl-1">
								<CollectionBranch
									folders={collection.folders}
									requests={collection.requests}
								/>
							</div>
						)}
					</div>
				);
			})}
		</div>
	);
}
