import { cn } from "#lib/utils";
import { methodTint } from "#lib/methodTint";
import { AuthPanel } from "../../features";
import { CollectionTree } from "../CollectionTree";
import { useWorkspaceStore, type WorkspaceView } from "../../stores";
import { useHistoryStore } from "../../stores/useHistoryStore";
import { useTestStore } from "../../stores/useTestStore";
import { useRealtimeRecentsStore } from "../../stores/useRealtimeRecentsStore";
import { useRealtimeStore } from "../../stores/useRealtimeStore";

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
		const id = fresh ? `test-new-${Date.now()}` : `test-${path}`;
		if (fresh) newTab(id);
		else void openPath(id, path ?? "");
		openTab({ id, title, kind: "test", filePath: path ?? undefined });
		requestView("requests");
	};

	return (
		<>
			<CollectionTree />
			{tests.length > 0 && (
				<div className="border-t border-border px-2 py-2">
					<SectionLabel>Tests</SectionLabel>
					<ul className="flex flex-col gap-0.5">
						{tests.map((t) => (
							<li key={t.path}>
								<button
									type="button"
									className="w-full truncate rounded px-2 py-0.5 text-left text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
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
			<AuthPanel />
		</>
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

/** Blurb shown for tools whose working set lives entirely in the main pane. */
function ToolContext({ description }: { description: string }) {
	return (
		<p className="px-3 pt-3 text-xs leading-relaxed text-muted-foreground">
			{description}
		</p>
	);
}

const TOOL_BLURBS = {
	mocks:
		"Stand-in servers that answer while the real API is down or unfinished.",
	diff: "Compare two OpenAPI documents and see what changed for consumers.",
	jwt: "Paste a token to decode its claims, timing, and expiry.",
	graphql: "Introspect a GraphQL endpoint and browse its schema.",
	grpc: "Call gRPC methods against a reflected or protoset-backed server.",
	runners: "Run collections, paginate through lists, or fire bulk requests.",
	explorer: "Browse any OpenAPI document as a navigable reference.",
	docs: "Generate REST documentation from the workspace collections.",
	"spec-editor": "Edit the OpenAPI spec with live tree navigation.",
} satisfies Partial<Record<WorkspaceView, string>>;

function RealtimeRecents({ kind }: { kind: "ws" | "sse" }) {
	const recents = useRealtimeRecentsStore((s) => s.recents.filter((r) => r.kind === kind).slice(0, 12));
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

export function ContextSidebar({ className }: { className?: string }) {
	const activeView = useWorkspaceStore((s) => s.activeView);

	if (activeView === "home") return null;
	if (activeView === "websocket")
		return <aside className={cn("flex h-full w-full flex-col overflow-y-auto border-r border-border bg-card/30", className)}><RealtimeRecents kind="ws" /></aside>;
	if (activeView === "sse")
		return <aside className={cn("flex h-full w-full flex-col overflow-y-auto border-r border-border bg-card/30", className)}><RealtimeRecents kind="sse" /></aside>;
	if (activeView === "settings")
		return <aside className={cn("flex h-full w-full flex-col overflow-y-auto border-r border-border bg-card/30", className)}><p className="px-3 pt-3 text-xs text-muted-foreground">Settings — see main pane.</p></aside>;

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
		default:
			content = (
				<ToolContext
					description={
						TOOL_BLURBS[activeView] ?? "Select a tool from the rail to begin."
					}
				/>
			);
	}

	return (
		<aside
			className={cn(
				"flex h-full w-full flex-col overflow-y-auto border-r border-border bg-card/30",
				className,
			)}
		>
			{content}
		</aside>
	);
}
