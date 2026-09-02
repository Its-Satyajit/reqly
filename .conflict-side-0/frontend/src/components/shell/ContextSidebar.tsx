import { cn } from "#lib/utils";
import { methodTint } from "#lib/methodTint";
import {
	Palette,
	FolderGit2,
	Database,
	Globe,
	Lock,
	Terminal,
	Keyboard,
	Info,
} from "lucide-react";
import { AuthPanel } from "../../features/auth-panel/AuthPanel";
import { CollectionTree } from "../CollectionTree";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";
import { useHistoryStore } from "../../stores/useHistoryStore";
import { useTestStore } from "../../stores/useTestStore";
import { useRealtimeRecentsStore } from "../../stores/useRealtimeRecentsStore";
import { useRealtimeStore } from "../../stores/useRealtimeStore";
import { useMockStore } from "../../stores/useMockStore";
import { useDocsStore } from "../../stores/useDocsStore";
import { useSpecEditorStore } from "../../stores/useSpecEditorStore";
import { useSettingsStore, type SettingsTabId } from "../../stores/useSettingsStore";

function SectionLabel({ children }: { children: React.ReactNode }) {
	return (
		<p className="px-2 pb-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
			{children}
		</p>
	);
}

function RequestsContext() {
	const tests = useTestStore((s) => s.tests);
	const openPath = useTestStore((s) => s.openPath);
	const newTab = useTestStore((s) => s.newTab);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const requestView = useWorkspaceStore((s) => s.requestView);

	const openTestTab = (path: string | null, title: string, fresh: boolean) => {
		const id = fresh ? `test-new-${crypto.randomUUID()}` : `test-${path}`;
		if (fresh) newTab(id);
		else void openPath(id, path ?? "");
		openTab({ id, title, kind: "test", filePath: path ?? undefined });
		requestView("requests");
	};

	return (
		<div className="flex flex-col">
			<CollectionTree />
			{tests.length > 0 && (
				<div className="border-t border-border px-2 py-2">
					<SectionLabel>Tests</SectionLabel>
					<ul className="flex flex-col gap-0.5">
						{tests.map((t) => (
							<li key={t.path}>
								<button
									type="button"
									className="w-full truncate rounded px-2 py-1 text-left text-xs text-muted-foreground transition-colors hover:bg-background hover:text-foreground"
									title={t.path}
									onClick={() => openTestTab(t.path, t.name || t.path, false)}
								>
									{t.name || t.path}
								</button>
							</li>
						))}
					</ul>
				</div>
			)}
			<div className="border-t border-border bg-muted/20">
				<details className="group" open>
					<summary className="flex cursor-pointer list-none items-center justify-between px-3 py-2 text-xs font-medium text-muted-foreground hover:text-foreground">
						<span className="font-mono text-[11px] uppercase tracking-wide">Auth</span>
						<span className="text-[11px] transition-transform group-open:rotate-180">▾</span>
					</summary>
					<div className="border-t border-border/60">
						<AuthPanel />
					</div>
				</details>
			</div>
		</div>
	);
}

function EnvironmentsContext() {
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);

	return (
		<div className="px-2 py-2">
			<SectionLabel>Environments</SectionLabel>
			{environments.length === 0 ? (
				<p className="px-2 pb-2 text-xs text-muted-foreground">
					No environments yet. Create one to switch variable sets.
				</p>
			) : (
				<ul className="flex flex-col gap-0.5">
					{environments.map((env) => {
						const active = env.id === activeEnvironmentId;
						return (
							<li key={env.id}>

								<span
									className={cn(
										"flex items-center gap-2 truncate rounded px-2 py-0.5 text-xs",
										active
											? "bg-muted font-medium text-foreground"
											: "text-muted-foreground",
									)}
								>
									<span
										aria-hidden
										className={cn(
											"size-1.5 shrink-0 rounded-full",
											active ? "bg-primary" : "bg-transparent",
										)}
									/>
									<span className="truncate">{env.name}</span>
								</span>
							</li>
						);
					})}
				</ul>
			)}
		</div>
	);
}

function HistoryContext() {
	const pool = useHistoryStore((s) => s.pool);
	const requestView = useWorkspaceStore((s) => s.requestView);
	const recent = pool.slice(0, 12);

	return (
		<div className="px-2 py-2">
			<SectionLabel>Recent</SectionLabel>
			{recent.length === 0 ? (
				<p className="px-2 pb-2 text-xs text-muted-foreground">
					Send a request and it will show up here.
				</p>
			) : (
				<ul className="flex flex-col gap-0.5">
					{recent.map((entry) => (
						<li key={entry.id}>
							<button
								type="button"
								onClick={() => requestView("history")}
								className="flex w-full items-center gap-1.5 truncate rounded px-2 py-0.5 text-left text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
								title={`${entry.method} ${entry.url}`}
							>
								<span
									className={cn(
										"w-10 shrink-0 text-left text-[11px] font-medium",
										// SAFETY: unknown methods fall through to the muted tint below.
									methodTint[entry.method as keyof typeof methodTint] ?? "text-muted-foreground",
									)}
								>
									{entry.method}
								</span>
								<span className="truncate">{entry.requestPath || entry.url}</span>
							</button>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}

function MocksContext() {
	const routes = useMockStore((s) => s.routes);
	const scenarios = useMockStore((s) => s.scenarios);
	const status = useMockStore((s) => s.status);
	const port = useMockStore((s) => s.port);
	const activeScenarioId = useMockStore((s) => s.activeScenarioId);
	const setActiveScenario = useMockStore((s) => s.setActiveScenario);
	const addRoute = useMockStore((s) => s.addRoute);

	return (
		<div className="flex flex-col gap-3 p-2">
			<div className="flex items-center justify-between border-b border-border/60 pb-2 px-1">
				<div className="flex items-center gap-1.5 font-mono text-[11px]">
					<span
						className={cn(
							"size-2 rounded-full",
							status.running ? "bg-status-ok animate-pulse" : "bg-muted-foreground/40",
						)}
					/>
					<span className="font-semibold">{status.running ? `Port :${port}` : "Offline"}</span>
				</div>
				<button
					type="button"
					onClick={addRoute}
					className="rounded border border-border/80 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground hover:bg-muted hover:text-foreground"
				>
					+ Route
				</button>
			</div>

			<div>
				<SectionLabel>Routes ({routes.length})</SectionLabel>
				{routes.length === 0 ? (
					<p className="px-2 text-xs text-muted-foreground">No mock routes defined.</p>
				) : (
					<ul className="flex flex-col gap-0.5">
						{routes.map((r) => {
							// SAFETY: Unknown mock HTTP methods fallback to muted tint
							const tintClass = methodTint[r.method as keyof typeof methodTint] ?? "text-muted-foreground";
							const routeKey = r.id || `${r.method}-${r.path}`;
							return (
								<li key={routeKey} className="flex items-center gap-1.5 rounded px-2 py-1 text-xs hover:bg-muted/40 font-mono">
									<span className={cn("text-[10px] font-bold", tintClass)}>
										{r.method}
									</span>
									<span className="truncate text-foreground/90">{r.path}</span>
									<span className="ml-auto text-[10px] text-muted-foreground">{r.status}</span>
								</li>
							);
						})}
					</ul>
				)}
			</div>

			{scenarios.length > 0 && (
				<div className="border-t border-border/60 pt-2">
					<SectionLabel>Scenarios ({scenarios.length})</SectionLabel>
					<ul className="flex flex-col gap-0.5">
						{scenarios.map((sc) => (
							<li key={sc.id}>
								<button
									type="button"
									onClick={() => setActiveScenario(sc.id === activeScenarioId ? null : sc.id)}
									className={cn(
										"flex w-full items-center justify-between rounded px-2 py-1 text-xs font-mono transition-colors",
										sc.id === activeScenarioId
											? "bg-primary/10 font-semibold text-primary"
											: "text-muted-foreground hover:bg-muted hover:text-foreground",
									)}
								>
									<span className="truncate">{sc.name}</span>
									<span className="text-[10px] opacity-70">{sc.routes?.length ?? 0} routes</span>
								</button>
							</li>
						))}
					</ul>
				</div>
			)}
		</div>
	);
}

function DocsContext() {
	const collections = useWorkspaceStore((s) => s.workspaceTree?.collections ?? []);
	const selected = useDocsStore((s) => s.selected);
	const toggleCollection = useDocsStore((s) => s.toggleCollection);
	const result = useDocsStore((s) => s.result);
	const activeFile = useDocsStore((s) => s.activeFile);
	const setActiveFile = useDocsStore((s) => s.setActiveFile);
	const selectedSet = new Set(selected);

	return (
		<div className="flex flex-col gap-3 p-2">
			<div>
				<SectionLabel>Collections to Document</SectionLabel>
				<ul className="flex flex-col gap-0.5">
					{collections.map((c) => {
						const isSelected = selected.length === 0 || selectedSet.has(c.name);
						return (
							<li key={c.name}>
								<label className="flex items-center gap-2 rounded px-2 py-1 text-xs hover:bg-muted/40 cursor-pointer">
									<input
										type="checkbox"
										checked={isSelected}
										onChange={() => toggleCollection(c.name)}
										className="size-3.5 rounded border-border accent-primary"
									/>
									<span className="truncate font-mono">{c.name}</span>
								</label>
							</li>
						);
					})}
				</ul>
			</div>

			{result && result.files.length > 0 && (
				<div className="border-t border-border/60 pt-2">
					<SectionLabel>Generated Files ({result.files.length})</SectionLabel>
					<ul className="flex flex-col gap-0.5">
						{result.files.map((f) => (
							<li key={f.name}>
								<button
									type="button"
									onClick={() => setActiveFile(f.name)}
									className={cn(
										"w-full truncate rounded px-2 py-1 text-left font-mono text-xs transition-colors",
										activeFile === f.name
											? "bg-primary/10 font-medium text-primary"
											: "text-muted-foreground hover:bg-muted hover:text-foreground",
									)}
								>
									{f.name}
								</button>
							</li>
						))}
					</ul>
				</div>
			)}
		</div>
	);
}

function SpecEditorContext() {
	const filePath = useSpecEditorStore((s) => s.filePath);
	const diagnostics = useSpecEditorStore((s) => s.diagnostics);
	const dirty = useSpecEditorStore((s) => s.dirty);
	const selectedId = useSpecEditorStore((s) => s.selectedId);
	const setSelected = useSpecEditorStore((s) => s.setSelected);

	return (
		<div className="flex flex-col gap-3 p-2">
			<div className="border-b border-border/60 pb-2 px-1">
				<p className="font-mono text-[11px] font-semibold text-foreground truncate">{filePath}</p>
				<p className="font-mono text-[10px] text-muted-foreground">
					{dirty ? "Unsaved changes" : "Synchronized"} · {diagnostics.length} issues
				</p>
			</div>

			<div>
				<SectionLabel>Navigation Quick Links</SectionLabel>
				<ul className="flex flex-col gap-0.5 font-mono text-xs">
					{[
						{ id: "info", label: "API Info & Metadata" },
						{ id: "servers", label: "Target Servers" },
						{ id: "paths", label: "Path Endpoints" },
						{ id: "components", label: "Components & Schemas" },
					].map((sec) => (
						<li key={sec.id}>
							<button
								type="button"
								onClick={() => setSelected(sec.id)}
								className={cn(
									"w-full rounded px-2 py-1 text-left transition-colors",
									selectedId === sec.id
										? "bg-primary/10 font-semibold text-primary"
										: "text-muted-foreground hover:bg-muted hover:text-foreground",
								)}
							>
								{sec.label}
							</button>
						</li>
					))}
				</ul>
			</div>
		</div>
	);
}

function DiffContext() {
	return (
		<div className="flex flex-col gap-3 p-2">
			<SectionLabel>OpenAPI Compare</SectionLabel>
			<p className="px-2 text-xs leading-relaxed text-muted-foreground font-mono">
				Compare live endpoints, revisions, and OpenAPI specifications to detect breaking changes before shipping.
			</p>
		</div>
	);
}

function RunnersContext() {
	const collections = useWorkspaceStore((s) => s.workspaceTree?.collections ?? []);
	const requestView = useWorkspaceStore((s) => s.requestView);

	return (
		<div className="flex flex-col gap-3 p-2">
			<SectionLabel>Collections ({collections.length})</SectionLabel>
			<ul className="flex flex-col gap-0.5 font-mono text-xs">
				{collections.map((c) => (
					<li key={c.name}>
						<button
							type="button"
							onClick={() => requestView("runners")}
							className="w-full truncate rounded px-2 py-1 text-left text-muted-foreground hover:bg-muted hover:text-foreground"
						>
							▶ {c.name}
						</button>
					</li>
				))}
			</ul>
		</div>
	);
}

function RealtimeRecents({ kind }: { kind: "ws" | "sse" }) {
	const allRecents = useRealtimeRecentsStore((s) => s.recents);
	const recents = allRecents.filter((r) => r.kind === kind).slice(0, 12);
	const update = useRealtimeStore((s) => s.update);
	const pageId = kind === "ws" ? "realtime-websocket-page" : "realtime-sse-page";
	if (recents.length === 0) {
		return <p className="px-3 pt-3 text-xs leading-relaxed text-muted-foreground">Connect to an endpoint and it will show up here.</p>;
	}
	return (
		<div className="px-2 py-2">
			<SectionLabel>Recent</SectionLabel>
			<ul className="flex flex-col gap-0.5">
				{recents.map((r) => (
					<li key={r.url}>
						<button type="button" onClick={() => update(pageId, { url: r.url })} className="w-full truncate rounded px-2 py-0.5 text-left text-xs text-muted-foreground hover:bg-muted hover:text-foreground" title={r.url}>{r.url}</button>
					</li>
				))}
			</ul>
		</div>
	);
}

function SettingsContext() {
	const activeTab = useSettingsStore((s) => s.activeTab);
	const setActiveTab = useSettingsStore((s) => s.setActiveTab);

	const items: { id: SettingsTabId; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
		{ id: "appearance", label: "Appearance", icon: Palette },
		{ id: "workspace", label: "Workspace", icon: FolderGit2 },
		{ id: "storage", label: "Storage & Retention", icon: Database },
		{ id: "network", label: "Network & Proxy", icon: Globe },
		{ id: "security", label: "TLS & Security", icon: Lock },
		{ id: "cicd", label: "CI / CD", icon: Terminal },
		{ id: "shortcuts", label: "Shortcuts", icon: Keyboard },
		{ id: "about", label: "About", icon: Info },
	];

	return (
		<div className="flex flex-col gap-1 p-2">
			<SectionLabel>Settings</SectionLabel>
			<ul className="flex flex-col gap-1 mt-1">
				{items.map((item) => {
					const Icon = item.icon;
					const selected = activeTab === item.id;
					return (
						<li key={item.id}>
							<button
								type="button"
								onClick={() => setActiveTab(item.id)}
								className={cn(
									"flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-xs font-mono transition-colors text-left",
									selected
										? "bg-primary/10 text-primary font-semibold border border-primary/30"
										: "text-muted-foreground hover:bg-muted/70 hover:text-foreground border border-transparent",
								)}
							>
								<Icon className={cn("size-3.5 shrink-0", selected ? "text-primary" : "text-muted-foreground")} />
								<span className="truncate">{item.label}</span>
							</button>
						</li>
					);
				})}
			</ul>
		</div>
	);
}

export function ContextSidebar({ className }: { className?: string }) {
	const activeView = useWorkspaceStore((s) => s.activeView);

	if (activeView === "home") return null;

	let content: React.ReactNode;
	switch (activeView) {
		case "requests":
			content = <RequestsContext />;
			break;
		case "environments":
			content = <EnvironmentsContext />;
			break;
		case "history":
			content = <HistoryContext />;
			break;
		case "mocks":
			content = <MocksContext />;
			break;
		case "docs":
			content = <DocsContext />;
			break;
		case "spec-editor":
			content = <SpecEditorContext />;
			break;
		case "diff":
			content = <DiffContext />;
			break;
		case "runners":
			content = <RunnersContext />;
			break;
		case "websocket":
			content = <RealtimeRecents kind="ws" />;
			break;
		case "sse":
			content = <RealtimeRecents kind="sse" />;
			break;
		case "settings":
			content = <SettingsContext />;
			break;
		default:
			content = (
				<div className="p-3">
					<SectionLabel>{activeView}</SectionLabel>
					<p className="px-2 text-xs leading-relaxed text-muted-foreground">
						Select a tool from the rail to begin.
					</p>
				</div>
			);
	}

	return (
		<aside
			className={cn(
				"flex h-full w-full flex-col overflow-y-auto border-r border-border bg-card",
				className,
			)}
		>
			{content}
		</aside>
	);
}
