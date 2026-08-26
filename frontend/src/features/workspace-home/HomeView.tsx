import { useEffect } from "react";
import {
	ArrowUpRight,
	FileDown,
	Plus,
	SquareArrowOutDownLeft,
	FolderOpen,
} from "lucide-react";
import { cn } from "#lib/utils";
import { Button } from "../../components/ui/button";
import { ImportDialog, ExportDialog } from "../../features";
import { useWorkspaceStore } from "../../stores";
import { useWorkspaceBootstrapStore } from "../../stores/useWorkspaceBootstrap";
import { useImportStore, useExportStore } from "../../stores";
import { useHistoryStore } from "../../stores/useHistoryStore";
import { NEW_REQUEST_TAB_ID } from "../../stores/useRequestStore";

const methodTint = {
	GET: "text-method-get",
	POST: "text-method-post",
	PUT: "text-method-put",
	PATCH: "text-method-put",
	DELETE: "text-method-delete",
} satisfies Record<string, string>;

function Stat({ value, label }: { value: number | string; label: string }) {
	return (
		<div className="flex min-w-24 flex-col gap-0.5 px-4 py-3">
			<span className="font-data text-xl leading-tight">{value}</span>
			<span className="text-[11px] uppercase tracking-wider text-muted-foreground">
				{label}
			</span>
		</div>
	);
}

export function HomeView() {
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const collections = useWorkspaceStore((s) => s.workspaceTree?.collections);
	const environments = useWorkspaceStore((s) => s.environments);
	const requestView = useWorkspaceStore((s) => s.requestView);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const switchWorkspace = useWorkspaceBootstrapStore((s) => s.openFolder);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const historyPool = useHistoryStore((s) => s.pool);
	const loadPool = useHistoryStore((s) => s.loadPool);
	const poolLoaded = useHistoryStore((s) => s.poolLoaded);

	useEffect(() => {
		if (!poolLoaded) void loadPool();
	}, [poolLoaded, loadPool]);

	const newRequest = () => {
		openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		requestView("requests");
	};
	const openCollection = () => {
		requestView("requests");
	};

	const today = new Date().toDateString();
	const requestsToday = historyPool.filter(
		(e) => new Date(e.createdAt).toDateString() === today,
	).length;
	const recent = historyPool.slice(0, 8);

	const collectionCount = collections?.length ?? 0;
	const envCount = environments.length;
	const requestCount = historyPool.length;
	const hasData = collectionCount > 0 || envCount > 0 || requestCount > 0;

	return (
		<section className="h-full min-h-0 overflow-y-auto">
			<div className="mx-auto flex w-full max-w-3xl flex-col gap-8 p-6 pb-16 sm:p-10">
				{/* Hero: the workspace rendered as the request line it really is. */}
				<header className="flex flex-col gap-1 pt-4">
					<p className="font-data text-xs text-muted-foreground">
						GET reqly://workspace/
						<span className="text-foreground">{workspaceName ?? "default"}</span>
					</p>
					<h1 className="text-2xl font-semibold tracking-tight">
						{hasData ? "Welcome back" : "Welcome to Reqly"}
					</h1>
				</header>

				{hasData ? (
					<div className="flex flex-wrap items-stretch divide-x divide-border rounded-lg border border-border bg-card">
						<Stat value={requestCount} label="Requests" />
						<Stat value={envCount} label="Environments" />
						<Stat value={collectionCount} label="Collections" />
						<Stat value={requestsToday} label="Recent Activity" />
					</div>
				) : (
					<div className="rounded-lg border border-dashed border-border p-8 text-center">
						<p className="mb-4 text-sm text-muted-foreground">
							Get started by creating your first request, importing an API spec, or setting up an environment.
						</p>
						<div className="flex flex-wrap justify-center gap-2">
							<Button onClick={newRequest}>
								<Plus className="size-4" aria-hidden />
								Create your first request
							</Button>
							<Button variant="outline" onClick={() => setImportOpen(true)}>
								<SquareArrowOutDownLeft className="size-4" aria-hidden />
								Import an API spec
							</Button>
							<Button variant="outline" onClick={() => requestView("environments")}>
								<Plus className="size-4" aria-hidden />
								Set up an environment
							</Button>
						</div>
					</div>
				)}

				<nav aria-label="Quick actions" className="flex flex-wrap gap-2">
					<Button onClick={newRequest}>
						<Plus className="size-4" aria-hidden />
						New request
					</Button>
					<Button variant="outline" onClick={() => setImportOpen(true)}>
						<SquareArrowOutDownLeft className="size-4" aria-hidden />
						Import API
					</Button>
					<Button variant="outline" onClick={openCollection}>
						<FolderOpen className="size-4" aria-hidden />
						Open collections
					</Button>
					<Button variant="outline" onClick={() => requestView("environments")}>
						<Plus className="size-4" aria-hidden />
						New environment
					</Button>
					<Button variant="ghost" onClick={() => void switchWorkspace()}>
						Switch workspace
					</Button>
					<Button variant="ghost" onClick={() => setExportOpen(true)}>
						<FileDown className="size-4" aria-hidden />
						Export
					</Button>
				</nav>

				<section aria-labelledby="recent-requests">
					<h2
						id="recent-requests"
						className="pb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
					>
						Recent requests
					</h2>
					{recent.length === 0 ? (
						<div className="rounded-lg border border-dashed border-border p-8 text-center">
							<p className="text-sm text-muted-foreground">
								Nothing sent yet. Open a request and press Send.
							</p>
							<Button variant="link" onClick={newRequest}>
								Create your first request
								<ArrowUpRight className="size-3.5" aria-hidden />
							</Button>
						</div>
					) : (
						<ul className="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
							{recent.map((entry) => (
								<li key={entry.id}>
									<button
										type="button"
										onClick={() => requestView("history")}
										className="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-muted/60"
									>

										<span
											className={cn(
												"w-12 shrink-0 font-data text-xs font-medium",
												// SAFETY: unknown methods fall through to the muted tint below.
									methodTint[entry.method as keyof typeof methodTint] ?? "text-muted-foreground",
											)}
										>
											{entry.method}
										</span>
										<span className="min-w-0 flex-1 truncate font-data text-sm">
											{entry.requestPath || entry.url}
										</span>
										<span className="shrink-0 font-data text-xs tabular-nums text-muted-foreground">
											{entry.status}
										</span>
										<span className="w-16 shrink-0 text-right font-data text-xs tabular-nums text-muted-foreground">
											{entry.durationMs} ms
										</span>
									</button>
								</li>
							))}
						</ul>
					)}
				</section>
			</div>
			<ImportDialog onImported={() => void loadPool()} />
			<ExportDialog />
		</section>
	);
}
