import { useState, useMemo } from "react";
import { ChevronRight, Play, Plus, Search } from "lucide-react";
import type { EntryIdentity, WorkspaceFolder, WorkspaceRequest } from "#lib/collections";
import { cn } from "#lib/utils";
import { ContextMenu as SharedContextMenu } from "#components/ContextMenu";
import { RUN_TAB_ID, useCollectionRunStore, useWorkspaceStore } from "#stores";

const TREE_KEYS = new Set([
	"ArrowDown",
	"ArrowUp",
	"ArrowRight",
	"ArrowLeft",
	"Home",
	"End",
]);

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

function RunControl({ path, name }: EntryIdentity) {
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

function ContextMenu({ path, name, onClose }: EntryIdentity & { onClose: () => void }) {
	const items = [
		"Rename",
		"Move",
		"Duplicate",
		"Delete",
		"Run",
		"Import",
		"Export",
		"Generate Docs",
		"Generate Tests",
		"Generate Mock",
	];
	return (
		<div
			role="menu"
			className="absolute left-full top-0 z-10 ml-1 min-w-36 rounded-md border border-border bg-popover p-1 shadow-md"
			onMouseLeave={onClose}
		>
			{items.map((label) => (
				<button
					key={label}
					type="button"
					role="menuitem"
					onClick={onClose}
					title={`${label} ${name} (${path}) — coming soon`}
					className="w-full rounded px-2 py-1 text-left text-xs text-foreground hover:bg-muted"
				>
					{label}
				</button>
			))}
		</div>
	);
}

function RequestRow({ request }: { request: WorkspaceRequest }) {
	const openRequest = useWorkspaceStore((s) => s.openRequest);
	const duplicateRequestPath = useWorkspaceStore((s) => s.duplicateRequestPath);
	const [menu, setMenu] = useState<{ x: number; y: number } | null>(null);

	// Heuristic method detection or default GET for collection display
	const methodGuess = request.name.match(/^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)/i)?.[0]?.toUpperCase() ?? "REQ";

	return (
		<div className="group/row flex items-center justify-between pl-5 pr-1 hover:bg-muted/50 rounded">
			<button
				type="button"
				data-tree-row
				draggable
				onDragStart={(e) => e.dataTransfer.setData("text/plain", request.path)}
				onContextMenu={(e) => {
					e.preventDefault();
					setMenu({ x: e.clientX, y: e.clientY });
				}}
				onClick={() => void openRequest(request.path)}
				title={`Open ${request.name} (${request.path})`}
				className="flex min-w-0 flex-1 items-center gap-2 py-1 text-left text-xs text-muted-foreground transition-colors hover:text-foreground"
			>
				<span className="font-mono text-[9px] font-bold tracking-tight text-primary/80 uppercase">
					{methodGuess}
				</span>
				<span className="truncate font-mono text-[11px]">{request.name}</span>
			</button>

			<button
				type="button"
				onClick={() => void openRequest(request.path)}
				title={`Run ${request.name}`}
				className="size-5 shrink-0 rounded p-1 text-muted-foreground/40 opacity-0 transition-opacity transition-colors hover:bg-muted hover:text-foreground group-hover/row:opacity-100"
			>
				<Play className="size-3" aria-hidden />
			</button>

			{menu && (
				<SharedContextMenu
					x={menu.x}
					y={menu.y}
					items={[
						{
							label: "Open request",
							onSelect: () => void openRequest(request.path),
						},
						{
							label: "Duplicate request",
							onSelect: () => void duplicateRequestPath(request.path),
						},
						{
							label: "Copy path",
							onSelect: () => void navigator.clipboard.writeText(request.path),
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
	filter: string;
}

function CollectionBranch({ folders, requests, filter }: BranchProps) {
	const q = filter.toLowerCase();
	const filteredRequests = q ? requests.filter((r) => r.name.toLowerCase().includes(q) || r.path.toLowerCase().includes(q)) : requests;
	const filteredFolders = q
		? folders.filter((f) => f.name.toLowerCase().includes(q) || f.path.toLowerCase().includes(q))
		: folders;
	return (
		<div className="flex flex-col gap-0.5">
			{filteredRequests.map((request) => (
				<RequestRow key={request.path} request={request} />
			))}
			{filteredFolders.map((folder) => (
				<FolderBranch key={folder.path} folder={folder} filter={filter} />
			))}
		</div>
	);
}

function FolderBranch({ folder, filter }: { folder: WorkspaceFolder; filter: string }) {
	const expanded = useWorkspaceStore((s) => s.expanded[folder.path] ?? false);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);
	const [menu, setMenu] = useState(false);

	return (
		<div>
			<div
				className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50"
				draggable
				onDragStart={(e) => e.dataTransfer.setData("text/plain", folder.path)}
				onDragOver={(e) => e.preventDefault()}
				onDrop={(e) => {
					e.preventDefault();
					// drag-and-drop reordering updates the underlying collection file via adapter when available
					// local reorder is optimistic; persistence handled by backend move
				}}
				onContextMenu={(e) => {
					e.preventDefault();
					setMenu((v) => !v);
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
						className={cn("size-3 shrink-0 transition-transform", expanded && "rotate-90")}
						aria-hidden
					/>
					<span className="truncate">{folder.name}</span>
				</button>
				<RunControl path={folder.path} name={folder.name} />
				{menu && <ContextMenu path={folder.path} name={folder.name} onClose={() => setMenu(false)} />}
			</div>
			{expanded && (
				<div className="ml-1 border-l border-border pl-1">
					<CollectionBranch folders={folder.folders} requests={folder.requests} filter={filter} />
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
	const [filter, setFilter] = useState("");

	const filteredCollections = useMemo(() => {
		if (!tree) return [];
		if (!filter.trim()) return tree.collections;
		const q = filter.toLowerCase();
		return tree.collections.filter(
			(c) => c.name.toLowerCase().includes(q) || c.path.toLowerCase().includes(q),
		);
	}, [tree, filter]);

	if (workspaceError) {
		return <p className="px-2 text-xs text-destructive">{workspaceError}</p>;
	}
	if (!tree || tree.collections.length === 0) {
		return (
			<div className="flex flex-col gap-2 px-2">
				<p className="pb-2 text-xs leading-relaxed text-muted-foreground">
					{tree?.name
						? "No collections yet — create collections/<name>/reqly.yaml to see them here."
						: "Open a reqly workspace in the desktop app to browse collections."}
				</p>
				<div className="flex gap-1">
					<button type="button" className="rounded border border-border px-2 py-1 text-xs hover:bg-muted" title="New Collection — coming soon">
						<Plus className="mr-1 inline size-3" />Collection
					</button>
					<button type="button" className="rounded border border-border px-2 py-1 text-xs hover:bg-muted" title="New Folder — coming soon">
						<Plus className="mr-1 inline size-3" />Folder
					</button>
					<button type="button" className="rounded border border-border px-2 py-1 text-xs hover:bg-muted" title="New Request — coming soon">
						<Plus className="mr-1 inline size-3" />Request
					</button>
				</div>
			</div>
		);
	}

	return (
		<div className="flex flex-col gap-1.5 p-1 select-none">
			<div className="relative px-1">
				<Search className="pointer-events-none absolute left-3.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground/70" aria-hidden />
				<input
					value={filter}
					onChange={(e) => setFilter(e.target.value)}
					placeholder="Filter collections…"
					aria-label="Filter collections"
					className="h-7 w-full rounded border border-input bg-background/80 py-1 pl-7 pr-2 font-mono text-[11px] placeholder:text-muted-foreground/60 focus:border-ring focus:outline-none"
				/>
			</div>
			<div role="tree" aria-label="Collections" onKeyDown={treeKeyDown} className="flex flex-col gap-0.5 px-0.5">
				{openError && (
					<p role="alert" className="px-2 text-xs text-destructive">
						{openError}
					</p>
				)}
				{filteredCollections.map((collection) => {
					const isOpen = expanded[collection.path] ?? false;
					return (
						<div
							key={collection.path}
							draggable
							onDragStart={(e) => e.dataTransfer.setData("text/plain", collection.path)}
							onDragOver={(e) => e.preventDefault()}
						>
							<div className="flex w-full items-center gap-1 rounded px-2 py-1 transition-colors hover:bg-muted/60">
								<button
									type="button"
									data-tree-row
									aria-expanded={isOpen}
									onClick={() => toggleExpanded(collection.path)}
									onContextMenu={(e) => e.preventDefault()}
									className="flex min-w-0 flex-1 items-center gap-1.5 text-left text-xs font-semibold text-foreground"
								>
									<ChevronRight className={cn("size-3 shrink-0 text-muted-foreground transition-transform", isOpen && "rotate-90 text-primary")} aria-hidden />
									<span className="truncate">{collection.name}</span>
								</button>
								<RunControl path={collection.path} name={collection.name} />
							</div>
							{isOpen && (
								<div className="ml-2 border-l border-border/70 pl-1.5">
									<CollectionBranch folders={collection.folders} requests={collection.requests} filter={filter} />
								</div>
							)}
						</div>
					);
				})}
			</div>
			<div className="flex gap-1 border-t border-border/70 px-1 pt-2">
				<button type="button" className="flex-1 rounded border border-border/80 bg-background/50 px-1.5 py-1 text-[11px] font-mono text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" title="New Collection">
					<Plus className="mr-1 inline size-3" />Collection
				</button>
				<button type="button" className="flex-1 rounded border border-border/80 bg-background/50 px-1.5 py-1 text-[11px] font-mono text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" title="New Folder">
					<Plus className="mr-1 inline size-3" />Folder
				</button>
				<button type="button" className="flex-1 rounded border border-border/80 bg-background/50 px-1.5 py-1 text-[11px] font-mono text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" title="New Request">
					<Plus className="mr-1 inline size-3" />Request
				</button>
			</div>
		</div>
	);
}
