import { ChevronRight, FolderPlus, Play, Search } from "lucide-react";
import { useState } from "react";
import type { WorkspaceFolder, WorkspaceRequest } from "#lib/collections";
import { methodTintClass } from "#lib/status";
import { cn } from "#lib/utils";
import { Button } from "#components/ui/button";
import { ContextMenu } from "#components/ContextMenu";
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
		<Button
			type="button"
			variant="ghost"
			size="icon-xs"
			onClick={run}
			disabled={running}
			title={running ? "A run is already in progress" : `Run ${name}`}
			aria-label={`Run ${name}`}
			className="shrink-0 text-muted-foreground/60 hover:text-status-ok disabled:opacity-40"
		>
			<Play className="size-3 fill-current" aria-hidden />
		</Button>
	);
}

/** MethodChip is the small method badge on request rows, tinted per the
 * GitHub REST-doc method ramp (GET green, POST amber, DELETE red). */
function MethodChip({ method }: { method: string }) {
	if (!method) return null;
	return (
		<span
			className={cn(
				"font-data shrink-0 rounded-full border border-border bg-muted/40 px-1.5 py-px text-2xs font-semibold uppercase",
				methodTintClass(method),
			)}
		>
			{method}
		</span>
	);
}

function requestMatches(request: WorkspaceRequest, query: string): boolean {
	return request.name.toLowerCase().includes(query);
}

function folderHasMatch(folder: WorkspaceFolder, query: string): boolean {
	if (folder.name.toLowerCase().includes(query)) return true;
	if (folder.requests.some((r) => requestMatches(r, query))) return true;
	return folder.folders.some((f) => folderHasMatch(f, query));
}

/** AddFolderControl opens the new-folder dialog scoped to a container. */
function AddFolderControl({
	path,
	name,
	onNewFolder,
}: {
	path: string;
	name: string;
	onNewFolder?: (path: string, name: string) => void;
}) {
	if (!onNewFolder) return null;
	return (
		<Button
			type="button"
			variant="ghost"
			size="icon-xs"
			onClick={() => onNewFolder(path, name)}
			title={`New folder in ${name}`}
			aria-label={`New folder in ${name}`}
			className="shrink-0 text-muted-foreground/60"
		>
			<FolderPlus className="size-3" aria-hidden />
		</Button>
	);
}

function RequestRow({
	request,
	filter,
}: {
	request: WorkspaceRequest;
	filter?: string;
}) {
	const openRequest = useWorkspaceStore((s) => s.openRequest);
	const duplicateRequestPath = useWorkspaceStore((s) => s.duplicateRequestPath);
	const [menu, setMenu] = useState<{ x: number; y: number } | null>(null);
	if (filter && !requestMatches(request, filter)) return null;
	return (
		<div style={{ paddingLeft: "1.5rem" }}>
			<button
				type="button"
				data-tree-row
				onClick={() => void openRequest(request.path)}
				onContextMenu={(e) => {
					e.preventDefault();
					setMenu({ x: e.clientX, y: e.clientY });
				}}
				title={`Open ${request.name}`}
				className="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-xs text-muted-foreground hover:bg-muted/50 hover:text-foreground"
			>
				<span className="size-3 shrink-0" aria-hidden />
				<MethodChip method={request.method ?? ""} />
				<span className="truncate">{request.name}</span>
			</button>
			{menu && (
				<ContextMenu
					x={menu.x}
					y={menu.y}
					items={[
						{
							label: "Duplicate request",
							onSelect: () => void duplicateRequestPath(request.path),
						},
					]}
					onClose={() => setMenu(null)}
				/>
			)}
		</div>
	);
}

interface BranchProps {
	folders: WorkspaceFolder[];
	requests: WorkspaceRequest[];
	filter?: string;
}

function CollectionBranch({ folders, requests, filter, onNewFolder }: BranchProps & { onNewFolder?: (path: string, name: string) => void }) {
	return (
		<div className="flex flex-col gap-0.5">
			{requests.map((request) => (
				<RequestRow key={request.path} request={request} filter={filter} />
			))}
			{folders.map((folder) => (
				<FolderBranch key={folder.path} folder={folder} filter={filter} onNewFolder={onNewFolder} />
			))}
		</div>
	);
}

function FolderBranch({ folder, filter, onNewFolder }: { folder: WorkspaceFolder; filter?: string; onNewFolder?: (path: string, name: string) => void }) {
	const expanded = useWorkspaceStore((s) => s.expanded[folder.path] ?? false);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);
	const [menu, setMenu] = useState<{ x: number; y: number } | null>(null);

	// While filtering, folders with matching descendants are shown expanded
	// regardless of their saved expansion state; non-matching folders hide.
	const forcedOpen = Boolean(filter) && folderHasMatch(folder, filter ?? "");
	if (filter && !forcedOpen) return null;
	const isOpen = expanded || forcedOpen;

	return (
		<div>
			<div
				className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50"
				onContextMenu={(e) => {
					if (!onNewFolder) return;
					e.preventDefault();
					setMenu({ x: e.clientX, y: e.clientY });
				}}
			>
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
				<AddFolderControl path={folder.path} name={folder.name} onNewFolder={onNewFolder} />
				<RunControl path={folder.path} name={folder.name} />
			</div>
			{menu && onNewFolder && (
				<ContextMenu
					x={menu.x}
					y={menu.y}
					items={[
						{
							label: "New folder",
							onSelect: () => onNewFolder(folder.path, folder.name),
						},
					]}
					onClose={() => setMenu(null)}
				/>
			)}
			{isOpen && (
				<div className="ml-1 border-l border-border pl-1">
					<CollectionBranch folders={folder.folders} requests={folder.requests} />
				</div>
			)}
		</div>
	);
}

export interface CollectionTreeProps {
	onNewCollection?: () => void;
	/** onNewFolder(path, name) opens the new-folder dialog for a container. */
	onNewFolder?: (path: string, name: string) => void;
}

export function CollectionTree({
	onNewCollection,
	onNewFolder,
}: CollectionTreeProps) {
	const tree = useWorkspaceStore((s) => s.workspaceTree);
	const workspaceError = useWorkspaceStore((s) => s.workspaceError);
	const openError = useWorkspaceStore((s) => s.openError);
	const expanded = useWorkspaceStore((s) => s.expanded);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);
	const [filter, setFilter] = useState("");
	const query = filter.trim().toLowerCase();

	if (workspaceError) {
		return <p className="px-2 text-xs text-destructive">{workspaceError}</p>;
	}
	if (!tree || tree.collections.length === 0) {
		return (
			<div className="flex flex-col gap-2 px-2 pb-4">
				<p className="text-xs leading-relaxed text-muted-foreground">
					{tree?.name
						? "No collections yet — create one to start organizing requests."
						: "Open a reqly workspace in the desktop app to browse collections."}
				</p>
				{tree?.name && onNewCollection && (
					<Button variant="outline" size="sm" onClick={onNewCollection}>
						New collection
					</Button>
				)}
			</div>
		);
	}

	return (
		<div
			role="tree"
			aria-label="Collections"
			onKeyDown={treeKeyDown}
			className="flex flex-col gap-0.5"
		>
			<div className="side-search mx-2 mb-1 flex shrink-0 items-center gap-1.5 rounded-full border border-input bg-background px-2.5 py-1 focus-within:border-primary/45">
				<Search className="size-3 shrink-0 text-muted-foreground" aria-hidden />
				<input
					value={filter}
					onChange={(e) => setFilter(e.target.value)}
					placeholder="Filter requests…"
					aria-label="Filter requests"
					spellCheck={false}
					className="min-w-0 flex-1 bg-transparent text-xs text-foreground outline-none placeholder:text-muted-foreground"
				/>
			</div>
			{openError && (
				<p role="alert" className="px-2 text-xs text-destructive">
					{openError}
				</p>
			)}
			{tree.collections.map((collection) => {
				const isOpen = expanded[collection.path] ?? false;
				const matchesSelf = collection.name.toLowerCase().includes(query);
				const forcedOpen = Boolean(query) && (matchesSelf || folderHasMatch(collection, query));
				if (query && !forcedOpen) return null;
				const showChildren = isOpen || Boolean(query);
				return (
					<div key={collection.path}>
			<div className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50">
							<button
								type="button"
								data-tree-row
								aria-expanded={showChildren}
								onClick={() => toggleExpanded(collection.path)}
								className="flex min-w-0 flex-1 items-center gap-1 text-left text-xs font-medium text-foreground"
							>
								<ChevronRight
									className={cn(
										"size-3 shrink-0 transition-transform",
										showChildren && "rotate-90",
									)}
									aria-hidden
								/>
								<span className="truncate">{collection.name}</span>
							</button>
							<AddFolderControl path={collection.path} name={collection.name} onNewFolder={onNewFolder} />
							<RunControl path={collection.path} name={collection.name} />
						</div>
						{showChildren && (
							<div className="ml-1 border-l border-border pl-1">
								<CollectionBranch
									folders={collection.folders}
									requests={collection.requests}
									filter={query || undefined}
									onNewFolder={onNewFolder}
								/>
							</div>
						)}
					</div>
				);
			})}
		</div>
	);
}
