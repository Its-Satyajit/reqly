import { useEffect, useMemo } from "react";
import {
	ArrowUpRight,
	Clock3,
	FileDown,
	FolderOpen,
	FolderPlus,
	History,
	Plus,
	SquareArrowOutDownLeft,
	Trash2,
	ChevronRight,
	Layers,
	Boxes,
	FlaskConical,
	Signal,
} from "lucide-react";
import { cn } from "#lib/utils";
import { methodTint } from "#lib/methodTint";
import { Button } from "../../components/ui/button";
import { ImportDialog } from "../import-dialog/ImportDialog";
import { ExportDialog } from "../export-dialog/ExportDialog";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";
import { useWorkspaceBootstrapStore } from "../../stores/useWorkspaceBootstrap";
import { useImportStore } from "../../stores/useImportStore";
import { useExportStore } from "../../stores/useExportStore";
import { useHistoryStore } from "../../stores/useHistoryStore";
import { NEW_REQUEST_TAB_ID } from "../../stores/useRequestStore";

function formatRelative(ts: number): string {
	const diff = Date.now() - ts;
	const m = Math.floor(diff / 60000);
	if (m < 1) return "now";
	if (m < 60) return `${m}m ago`;
	const h = Math.floor(m / 60);
	if (h < 24) return `${h}h ago`;
	const d = Math.floor(h / 24);
	if (d === 1) return "yesterday";
	if (d < 7) return `${d}d ago`;
	const date = new Date(ts);
	return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function formatDuration(ms: number): string {
	if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`;
	return `${ms}ms`;
}

function Stat({ icon: Icon, value, label, hint }: { icon: typeof Layers; value: number | string; label: string; hint?: string }) {
	return (
		<div className="flex min-w-0 flex-1 items-center gap-2.5 px-3 py-2.5">
			<span className="flex size-7 shrink-0 items-center justify-center rounded-md border border-border bg-muted/40">
				<Icon className="size-3.5 text-muted-foreground" aria-hidden />
			</span>
			<span className="min-w-0">
				<span className="flex items-baseline gap-1.5">
					<span className="font-mono text-[15px] font-semibold leading-none tabular-nums">{value}</span>
					<span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{label}</span>
				</span>
				{hint && <span className="block truncate font-mono text-[10px] leading-none text-muted-foreground/70">{hint}</span>}
			</span>
		</div>
	);
}

export function HomeView() {
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const workspacePath = useWorkspaceStore((s) => s.workspaceTree?.path);
	const collections = useWorkspaceStore((s) => s.workspaceTree?.collections);
	const environments = useWorkspaceStore((s) => s.environments);
	const requestView = useWorkspaceStore((s) => s.requestView);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const currentWorkspace = useWorkspaceStore((s) => s.currentWorkspace);
	const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
	const openDirect = useWorkspaceBootstrapStore((s) => s.openDirect);
	const recentWorkspaces = useWorkspaceBootstrapStore((s) => s.recentWorkspaces);
	const clearRecentWorkspaces = useWorkspaceBootstrapStore((s) => s.clearRecentWorkspaces);
	const setCreateModalOpen = useWorkspaceBootstrapStore((s) => s.setCreateModalOpen);
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

	const today = new Date().toDateString();
	const requestsToday = useMemo(
		() => historyPool.filter((e) => new Date(e.createdAt).toDateString() === today).length,
		[historyPool, today],
	);
	const recent = historyPool.slice(0, 6);

	const collectionCount = collections?.length ?? 0;
	const envCount = environments.length;
	const requestCount = historyPool.length;
	const hasData = collectionCount > 0 || envCount > 0 || requestCount > 0;

	const flatRequests = useMemo(() => {
		if (!collections) return [];
		const out: { name: string; path: string }[] = [];
		const walk = (folders: typeof collections) => {
			for (const c of folders) {
				for (const r of c.requests) out.push(r);
				const stack: typeof c.folders = [...c.folders];
				while (stack.length) {
					const f = stack.pop()!;
					for (const r of f.requests) out.push(r);
					stack.push(...f.folders);
				}
			}
		};
		walk(collections);
		return out.slice(0, 8);
	}, [collections]);

	return (
		<section className="flex h-full min-h-0 flex-col overflow-y-auto bg-background">
			<div className="mx-auto flex w-full max-w-[1280px] flex-col gap-4 p-4 pb-8 sm:p-5">
				{/* ── Dossier hero ────────────────────────────────────────────── */}
				<header className="relative overflow-hidden rounded-xl border border-border bg-card">
					{/* subtle top rule – coral accent as engineered honesty, not gradient */}
					<div className="h-[2px] w-full bg-primary" aria-hidden />
					<div className="flex flex-col gap-3 p-4 sm:p-5">
						<div className="flex flex-wrap items-start justify-between gap-3">
							<div className="min-w-0">
								<p className="flex items-center gap-1.5 font-mono text-[11px] leading-none tracking-wide text-muted-foreground">
									<span className="inline-flex size-1.5 rounded-full bg-status-ok shadow-[0_0_0_3px_color-mix(in_srgb,var(--status-ok)_18%,transparent)]" aria-hidden />
									<span className="select-all">reqly://workspace/</span>
									<span className="truncate font-medium text-foreground">{workspaceName ?? "default"}</span>
									<span className="hidden items-center gap-1 rounded bg-muted px-1 py-0.5 font-mono text-[10px] text-muted-foreground sm:inline-flex">
										<span className="size-1 rounded-full bg-status-ok" aria-hidden /> Local · Git-native
									</span>
								</p>
								<h1 className="mt-1.5 max-w-[32ch] text-balance text-[22px] font-semibold leading-none tracking-[-0.02em] sm:text-[26px]">
									{hasData ? (
										<>
											<span className="font-normal text-muted-foreground">Workspace</span> — {workspaceName ?? "Untitled"}
										</>
									) : (
										"Start a workspace that lives in your filesystem"
									)}
								</h1>
								<p
									className="mt-1.5 max-w-[64ch] truncate font-mono text-[11px] leading-none text-muted-foreground"
									title={workspacePath ?? currentWorkspace?.path ?? "—"}
								>
									{workspacePath ?? currentWorkspace?.path ?? "—"} · plain files · versionable with Git
								</p>
							</div>

							<div className="flex shrink-0 flex-wrap items-center gap-1.5">
								<Button size="sm" onClick={newRequest} className="gap-1.5">
									<Plus className="size-3.5" aria-hidden />
									New request
								</Button>
								<Button variant="outline" size="sm" onClick={() => setImportOpen(true)} className="gap-1.5">
									<SquareArrowOutDownLeft className="size-3.5" aria-hidden />
									Import
								</Button>
								<Button variant="outline" size="sm" onClick={() => requestView("environments")} className="gap-1.5">
									<FlaskConical className="size-3.5" aria-hidden />
									Environment
								</Button>
								<span className="hidden h-7 w-px bg-border sm:block" aria-hidden />
								<Button variant="ghost" size="sm" onClick={() => void openFolder()} className="gap-1.5">
									<FolderOpen className="size-3.5" aria-hidden />
									Open
								</Button>
								<Button variant="ghost" size="sm" onClick={() => setExportOpen(true)} className="gap-1.5">
									<FileDown className="size-3.5" aria-hidden />
									Export
								</Button>
							</div>
						</div>

						{/* compact metric strip – dense, not decorative */}
						<div className="flex flex-wrap items-stretch divide-x divide-border overflow-hidden rounded-lg border border-border bg-muted/20">
							<Stat icon={Signal} value={requestCount} label="requests" hint={`${requestsToday} today`} />
							<Stat icon={Layers} value={collectionCount} label="collections" hint={flatRequests.length ? `${flatRequests.length} files` : "no files yet"} />
							<Stat icon={Boxes} value={envCount} label="environments" hint={envCount ? environments.map((e) => e.name).slice(0, 2).join(" · ") : "default"} />
							<Stat icon={History} value={recentWorkspaces.length} label="recent workspaces" hint={recentWorkspaces[0]?.name ?? "—"} />
						</div>
					</div>
				</header>

				{/* ── Bento ─────────────────────────────────────────────────────── */}
				<div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.65fr)_minmax(300px,0.85fr)]">
					{/* LEFT */}
					<div className="flex min-w-0 flex-col gap-4">
						{/* Collections / structure preview */}
						<section aria-labelledby="workspace-structure" className="rounded-xl border border-border bg-card">
							<div className="flex items-center justify-between gap-2 border-b border-border px-3.5 py-2.5">
								<h2 id="workspace-structure" className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
									<Layers className="size-3.5" aria-hidden />
									Structure
									<span className="rounded bg-muted px-1 py-0.5 font-mono text-[10px] font-normal normal-case tracking-normal text-muted-foreground">
										{workspacePath?.split(/[\\/]/).pop() ?? workspaceName ?? "workspace"}/
									</span>
								</h2>
								<div className="flex items-center gap-1">
									<Button variant="ghost" size="xs" onClick={() => requestView("requests")} className="gap-1">
										Open collections
										<ArrowUpRight className="size-3" aria-hidden />
									</Button>
								</div>
							</div>

							{collectionCount === 0 ? (
								<div className="p-6 text-center">
									<p className="mx-auto max-w-[36ch] text-sm leading-snug text-muted-foreground">
										No collections yet. Create a request or import a spec — files land in <span className="font-mono text-foreground">collections/</span> as plain text you can commit.
									</p>
									<div className="mt-4 flex flex-wrap justify-center gap-2">
										<Button size="sm" onClick={newRequest} className="gap-1.5">
											<Plus className="size-3.5" aria-hidden /> New request
										</Button>
										<Button variant="outline" size="sm" onClick={() => setImportOpen(true)} className="gap-1.5">
											<SquareArrowOutDownLeft className="size-3.5" aria-hidden /> Import OpenAPI / cURL
										</Button>
									</div>
									<p className="mt-3 font-mono text-[11px] text-muted-foreground/70">collections/ · environments/ · tests/ · .reqly/</p>
								</div>
							) : (
								<div className="divide-y divide-border/60">
									{collections!.map((c) => {
										const folderCount = c.folders.length;
										const reqCount = c.requests.length + c.folders.reduce((a, f) => a + f.requests.length, 0);
										return (
											<div key={c.path} className="flex items-center gap-3 px-3.5 py-2.5">
												<span className="flex size-7 shrink-0 items-center justify-center rounded-md border border-border bg-background font-mono text-[11px] font-medium">
													{c.name.slice(0, 2).toUpperCase()}
												</span>
												<span className="min-w-0 flex-1">
													<span className="block truncate text-sm font-medium leading-none">{c.name}</span>
													<span className="block truncate font-mono text-[11px] leading-none text-muted-foreground">
														collections/{c.path || c.name} · {folderCount ? `${folderCount} folders · ` : ""}
														{reqCount} requests
													</span>
												</span>
												<Button variant="ghost" size="xs" onClick={() => requestView("requests")} aria-label={`Open ${c.name}`}>
													Open
												</Button>
											</div>
										);
									})}
									{flatRequests.length > 0 && (
										<div className="flex flex-wrap gap-1.5 px-3.5 py-2.5">
											{flatRequests.map((r) => (
												<button
													key={r.path}
													type="button"
													onClick={() => {
														requestView("requests");
														// open after nav – let the sidebar settle
														queueMicrotask(() => void useWorkspaceStore.getState().openRequest(r.path));
													}}
													className="inline-flex max-w-[18ch] items-center truncate rounded border border-border bg-muted/30 px-1.5 py-1 font-mono text-[11px] leading-none text-muted-foreground transition-colors hover:border-border hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
													title={r.path}
												>
													<span className="truncate">{r.name || r.path.split("/").pop()}</span>
												</button>
											))}
										</div>
									)}
								</div>
							)}
						</section>

						{/* Recent requests – dense ledger, not airy list */}
						<section aria-labelledby="recent-requests" className="overflow-hidden rounded-xl border border-border bg-card">
							<div className="flex items-center justify-between gap-2 border-b border-border px-3.5 py-2.5">
								<h2 id="recent-requests" className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
									<Clock3 className="size-3.5" aria-hidden />
									Recent requests
									<span className="font-mono text-[10px] font-normal normal-case tracking-normal text-muted-foreground/70">
										· from history.db (FTS5)
									</span>
								</h2>
								<div className="flex items-center gap-1">
									<span className="hidden font-mono text-[11px] tabular-nums text-muted-foreground sm:inline">
										{requestsToday} today
									</span>
									<Button variant="ghost" size="xs" onClick={() => requestView("history")} className="gap-1">
										View all
										<ChevronRight className="size-3" aria-hidden />
									</Button>
								</div>
							</div>

							{recent.length === 0 ? (
								<div className="p-6 text-center">
									<p className="text-sm text-muted-foreground">Nothing sent yet. Send a request and it lands here — local, indexed, searchable.</p>
									<Button variant="link" size="sm" onClick={newRequest} className="gap-1">
										Create your first request <ArrowUpRight className="size-3.5" aria-hidden />
									</Button>
								</div>
							) : (
								<ul className="divide-y divide-border/60">
									{recent.map((entry) => (
										<li key={entry.id}>
											<button
												type="button"
												onClick={() => requestView("history")}
												className="group flex w-full items-center gap-2.5 px-3.5 py-2 text-left transition-colors hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none"
											>
												<span
													className={cn(
														"w-[3.25rem] shrink-0 text-right font-mono text-[11px] font-semibold tracking-wide",
														// SAFETY: entry.method is a known HTTP method string from history; keyof check narrows to methodTint map.
														methodTint[entry.method as keyof typeof methodTint] ?? "text-muted-foreground",
													)}
												>
													{entry.method}
												</span>
												<span className="min-w-0 flex-1 truncate font-mono text-[12px] leading-none text-foreground/90 group-hover:text-foreground">
													{entry.requestPath || entry.url}
												</span>
												<span className="hidden shrink-0 items-center gap-1.5 sm:inline-flex">
													<span
														className={cn(
															"size-1.5 rounded-full",
															entry.status >= 200 && entry.status < 300
																? "bg-status-ok"
																: entry.status >= 300 && entry.status < 400
																	? "bg-status-redirect"
																	: entry.status >= 400 && entry.status < 500
																		? "bg-status-warn"
																		: "bg-status-error",
														)}
														aria-hidden
													/>
													<span className="font-mono text-[11px] tabular-nums text-muted-foreground">{entry.status}</span>
												</span>
												<span className="w-[4.5rem] shrink-0 text-right font-mono text-[11px] tabular-nums text-muted-foreground">
													{formatDuration(entry.durationMs)}
												</span>
												<ChevronRight
													className="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100"
													aria-hidden
												/>
											</button>
										</li>
									))}
								</ul>
							)}
						</section>
					</div>

					{/* RIGHT – workspace history ledger */}
					<div className="flex min-w-0 flex-col gap-4">
						<section aria-labelledby="workspace-history" className="overflow-hidden rounded-xl border border-border bg-card">
							<div className="flex items-center justify-between gap-2 border-b border-border px-3.5 py-2.5">
								<h2
									id="workspace-history"
									className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground"
								>
									<History className="size-3.5" aria-hidden />
									Workspace history
									<span className="rounded bg-primary/10 px-1 py-0.5 font-mono text-[10px] font-semibold normal-case tracking-normal text-primary">
										{recentWorkspaces.length}
									</span>
								</h2>
								{recentWorkspaces.length > 0 && (
									<button
										type="button"
										onClick={clearRecentWorkspaces}
										className="inline-flex items-center gap-1 rounded px-1.5 py-1 font-mono text-[11px] text-muted-foreground transition-colors hover:bg-muted hover:text-destructive focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
										title="Clear recent workspaces"
									>
										<Trash2 className="size-3" aria-hidden />
										Clear
									</button>
								)}
							</div>

							{/* current dossier */}
							<div className="border-b border-border bg-muted/20 px-3.5 py-3">
								<p className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">Current</p>
								<p className="mt-1 truncate text-sm font-semibold leading-none tracking-tight">{workspaceName ?? "Untitled workspace"}</p>
								<p className="mt-1 truncate font-mono text-[11px] leading-none text-muted-foreground" title={workspacePath ?? currentWorkspace?.path}>
									{workspacePath ?? currentWorkspace?.path ?? "—"}
								</p>
								<div className="mt-2 flex flex-wrap gap-1.5">
									<span className="inline-flex items-center gap-1 rounded border border-border bg-background px-1.5 py-1 font-mono text-[11px] leading-none">
										<span className="size-1.5 rounded-full bg-status-ok" aria-hidden /> active
									</span>
									<span className="inline-flex items-center rounded border border-border bg-background px-1.5 py-1 font-mono text-[11px] leading-none text-muted-foreground">
										{collectionCount} coll · {envCount} env
									</span>
								</div>
							</div>

							{recentWorkspaces.length === 0 ? (
								<div className="p-5">
									<p className="text-sm leading-snug text-muted-foreground">
										No previous workspaces yet. Opened workspaces are remembered here as file entries — click to jump back.
									</p>
									<div className="mt-3 flex flex-wrap gap-2">
										<Button size="sm" variant="outline" onClick={() => void openFolder()} className="gap-1.5">
											<FolderOpen className="size-3.5" aria-hidden />
											Open folder…
										</Button>
										<Button size="sm" variant="outline" onClick={() => setCreateModalOpen(true)} className="gap-1.5">
											<FolderPlus className="size-3.5" aria-hidden />
											Create workspace…
										</Button>
									</div>
									<p className="mt-3 font-mono text-[11px] text-muted-foreground/60">Stored locally · click any path to reopen</p>
								</div>
							) : (
								<ul className="divide-y divide-border/60" role="list">
									{recentWorkspaces.map((ws, idx) => {
										const isActive = currentWorkspace?.path === ws.path;
										const isMostRecent = idx === 0;
										return (
											<li key={ws.path} className="relative">
												{/* recency accent – coral for most recent, muted for rest */}
												<span
													className={cn(
														"pointer-events-none absolute inset-y-1 left-0 w-[2px] rounded-full",
														isActive ? "bg-primary" : isMostRecent ? "bg-primary/60" : "bg-border",
													)}
													aria-hidden
												/>
												<button
													type="button"
													onClick={() => {
														if (!isActive) void openDirect(ws.path);
													}}
													disabled={isActive}
													className={cn(
														"group flex w-full items-center gap-2.5 pl-3 pr-2.5 py-2.5 text-left transition-colors focus-visible:outline-none focus-visible:bg-muted/60",
														isActive ? "cursor-default bg-muted/30" : "hover:bg-muted/50 cursor-pointer",
													)}
													title={isActive ? "Current workspace" : `Open ${ws.path}`}
												>
													<span
														className={cn(
															"flex size-7 shrink-0 items-center justify-center rounded-md border text-[11px]",
															isActive
																? "border-primary/30 bg-primary/10 text-primary"
																: "border-border bg-muted/30 text-muted-foreground group-hover:border-border group-hover:bg-background group-hover:text-foreground",
														)}
														aria-hidden
													>
														<FolderOpen className="size-3.5" />
													</span>
													<span className="min-w-0 flex-1">
														<span className="flex items-center gap-1.5">
															<span className="truncate text-sm font-medium leading-none">{ws.name}</span>
															{isActive && (
																<span className="shrink-0 rounded bg-primary px-1 py-0.5 font-mono text-[9px] font-semibold uppercase leading-none tracking-wide text-primary-foreground">
																	active
																</span>
															)}
															{!isActive && isMostRecent && (
																<span className="shrink-0 rounded bg-muted px-1 py-0.5 font-mono text-[9px] uppercase leading-none tracking-wide text-muted-foreground">
																	last
																</span>
															)}
														</span>
														<span className="mt-0.5 block truncate font-mono text-[11px] leading-none text-muted-foreground">
															{ws.path}
														</span>
														<span className="mt-1 flex items-center gap-1 font-mono text-[10px] leading-none text-muted-foreground/70">
															<Clock3 className="size-2.5" aria-hidden />
															{formatRelative(ws.lastOpened)}
														</span>
													</span>
													{isActive ? (
														<span className="shrink-0 font-mono text-[11px] text-muted-foreground">—</span>
													) : (
														<span className="flex shrink-0 items-center gap-1 rounded border border-transparent px-1.5 py-1 font-mono text-[11px] leading-none text-muted-foreground transition-colors group-hover:border-border group-hover:bg-background group-hover:text-foreground">
															Open
															<ChevronRight className="size-3" aria-hidden />
														</span>
													)}
												</button>
											</li>
										);
									})}
								</ul>
							)}

							<div className="flex items-center gap-1.5 border-t border-border bg-muted/10 px-3.5 py-2.5">
								<Button variant="outline" size="sm" onClick={() => void openFolder()} className="flex-1 gap-1.5">
									<FolderOpen className="size-3.5" aria-hidden />
									Open folder…
								</Button>
								<Button variant="outline" size="sm" onClick={() => setCreateModalOpen(true)} className="flex-1 gap-1.5">
									<FolderPlus className="size-3.5" aria-hidden />
									Create…
								</Button>
							</div>
							<p className="border-t border-border/50 px-3.5 py-2 font-mono text-[10px] leading-snug text-muted-foreground/60">
								Workspaces are plain folders with <span className="text-muted-foreground">reqly.yaml</span> + <span className="text-muted-foreground">collections/</span>. Pick any previous path to reopen — tabs reset, envs reload.
							</p>
						</section>

						{/* Compact utilities / empty-state follow-up */}
						<section className="rounded-xl border border-dashed border-border bg-muted/10 p-3.5">
							<h3 className="text-xs font-semibold">Quick moves</h3>
							<p className="mt-1 font-mono text-[11px] leading-snug text-muted-foreground">
								This page is the dossiers + ledger. Collections live in the sidebar; history and bulk ops live in History.
							</p>
							<div className="mt-3 grid grid-cols-2 gap-1.5">
								<Button variant="outline" size="sm" onClick={() => requestView("history")} className="justify-start gap-1.5">
									<History className="size-3.5" aria-hidden /> History
								</Button>
								<Button variant="outline" size="sm" onClick={() => requestView("environments")} className="justify-start gap-1.5">
									<FlaskConical className="size-3.5" aria-hidden /> Environments
								</Button>
								<Button variant="outline" size="sm" onClick={() => setImportOpen(true)} className="justify-start gap-1.5">
									<SquareArrowOutDownLeft className="size-3.5" aria-hidden /> Import
								</Button>
								<Button variant="outline" size="sm" onClick={() => setExportOpen(true)} className="justify-start gap-1.5">
									<FileDown className="size-3.5" aria-hidden /> Export
								</Button>
							</div>
						</section>
					</div>
				</div>
			</div>

			<ImportDialog onImported={() => void loadPool()} />
			<ExportDialog />
		</section>
	);
}
