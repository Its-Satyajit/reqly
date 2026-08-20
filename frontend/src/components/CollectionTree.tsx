import type { WorkspaceFolder, WorkspaceRequest } from "#lib/collections";
import { cn } from "#lib/utils";
import { RUN_TAB_ID, useCollectionRunStore, useWorkspaceStore } from "#stores";

interface Props {
	folders: WorkspaceFolder[];
	requests: WorkspaceRequest[];
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
			className="shrink-0 rounded p-1 text-muted-foreground/60 hover:bg-muted/50 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
		>
			<svg className="size-3" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
				<path d="M8 5v14l11-7z" />
			</svg>
		</button>
	);
}

function RequestRow({ request }: { request: WorkspaceRequest }) {
	const openRequest = useWorkspaceStore((s) => s.openRequest);
	return (
		<button
			onClick={() => void openRequest(request.path)}
			className="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-xs text-muted-foreground hover:bg-muted/50 hover:text-foreground"
		>
			<span className="size-3 shrink-0 text-muted-foreground/60" aria-hidden>
				▸
			</span>
			<span className="truncate">{request.name}</span>
		</button>
	);
}

function FolderBranch({
	folder,
	depth,
}: {
	folder: WorkspaceFolder;
	depth: number;
}) {
	const expanded = useWorkspaceStore((s) => s.expanded[folder.path] ?? false);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

	return (
		<div>
			<div
				className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50"
				style={{ paddingLeft: `${0.5 + depth * 0.75}rem` }}
			>
				<button
					onClick={() => toggleExpanded(folder.path)}
					className="flex min-w-0 flex-1 items-center gap-1 text-left text-xs text-muted-foreground"
				>
					<span
						className={cn(
							"text-muted-foreground/60 transition-transform",
							expanded && "rotate-90",
						)}
						aria-hidden
					>
						▸
					</span>
					<span className="truncate">{folder.name}</span>
				</button>
				<RunControl path={folder.path} name={folder.name} />
			</div>
			{expanded && (
				<div className="ml-1 border-l border-border pl-1">
					<CollectionBranch
						folders={folder.folders}
						requests={folder.requests}
						depth={depth + 1}
					/>
				</div>
			)}
		</div>
	);
}

function CollectionBranch({
	folders,
	requests,
	depth,
}: Props & { depth: number }) {
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
	const openError = useWorkspaceStore((s) => s.openError);
	const expanded = useWorkspaceStore((s) => s.expanded);
	const toggleExpanded = useWorkspaceStore((s) => s.toggleExpanded);

	if (workspaceError) {
		return <p className="px-2 text-xs text-destructive">{workspaceError}</p>;
	}
	if (!tree || tree.collections.length === 0) {
		return (
			<p className="px-2 text-xs text-muted-foreground">
				{tree?.name
					? "No collections yet"
					: "Open a reqly workspace in the desktop app to browse collections."}
			</p>
		);
	}

	return (
		<div className="flex flex-col gap-0.5">
			{openError && (
				<p className="px-2 text-xs text-destructive">{openError}</p>
			)}
			{tree.collections.map((collection) => {
				const isOpen = expanded[collection.path] ?? false;
				return (
					<div key={collection.path}>
						<div className="flex w-full items-center gap-1 rounded-md px-2 py-1 hover:bg-muted/50">
							<button
								onClick={() => toggleExpanded(collection.path)}
								className="flex min-w-0 flex-1 items-center gap-1 text-left text-xs font-medium text-foreground"
							>
								<span
									className={cn(
										"text-muted-foreground/60 transition-transform",
										isOpen && "rotate-90",
									)}
									aria-hidden
								>
									▸
								</span>
								<span className="truncate">{collection.name}</span>
							</button>
							<RunControl path={collection.path} name={collection.name} />
						</div>
						{isOpen && (
							<div className="ml-1 border-l border-border pl-1">
								<CollectionBranch
									folders={collection.folders}
									requests={collection.requests}
									depth={1}
								/>
							</div>
						)}
					</div>
				);
			})}
		</div>
	);
}
