import type { ComponentType } from "react";
import {
	Clock,
	Diff,
	FileCode2,
	GitBranch,
	House,
	Import,
	Leaf,
	Play,
	Server,
	ShieldCheck,
	SquareTerminal,
	Wifi,
	type LucideIcon,
} from "lucide-react";
import { SiGraphql } from "@icons-pack/react-simple-icons";
import { ViewShell } from "../../components/shell/ViewLayout";
import { NEW_REQUEST_TAB_ID } from "#stores/useRequestStore";
import { useGitStore } from "#stores/useGitStore";
import type { HistoryEntry } from "#lib/history";
import { useHistoryStore } from "#stores/useHistoryStore";
import { type WorkspaceFolder, type WorkspaceTree } from "#lib/collections";
import { methodTintClass } from "#lib/status";
import { StatusPill } from "#components/status";
import { cn } from "#lib/utils";
import { useImportStore, useWorkspaceStore } from "#stores";

/** countTreeRequests totals requests across every collection/folder. */
function countTreeRequests(tree: WorkspaceTree | null): number {
	if (!tree) return 0;
	let count = 0;
	const walkFolder = (f: WorkspaceFolder) => {
		count += f.requests.length;
		for (const child of f.folders) walkFolder(child);
	};
	for (const c of tree.collections) {
		count += c.requests.length;
		for (const f of c.folders) {
			count += f.requests.length;
			for (const g of f.folders) walkFolder(g);
		}
	}
	return count;
}

/** relativeTime renders a compact "5m ago" style age for a history entry. */
function relativeTime(createdAt: string): string {
	const ms = Date.now() - new Date(createdAt).getTime();
	if (!Number.isFinite(ms)) return "";
	const minutes = Math.floor(ms / 60_000);
	if (minutes < 1) return "now";
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	return `${Math.floor(hours / 24)}d ago`;
}

/** isToday reports whether an entry was created today (local time). */
function isToday(createdAt: string): boolean {
	const d = new Date(createdAt);
	if (Number.isNaN(d.getTime())) return false;
	const now = new Date();
	return (
		d.getFullYear() === now.getFullYear() &&
		d.getMonth() === now.getMonth() &&
		d.getDate() === now.getDate()
	);
}

interface QuickAction {
	label: string;
	icon: ComponentType<{ className?: string }> | LucideIcon;
	onSelect: () => void;
}

/** QuickActionTile is one bordered square in the quick-actions grid. */
function QuickActionTile({ action }: { action: QuickAction }) {
	const Icon = action.icon;
	return (
		<button
			type="button"
			onClick={action.onSelect}
			className="flex flex-col items-center justify-center gap-1.5 rounded-lg border border-border bg-muted/20 px-2 py-3 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
		>
			<Icon className="size-4 text-primary" aria-hidden />
			{action.label}
		</button>
	);
}

/** Card is the shared bordered section container. */
function Card({
	title,
	children,
	className,
}: {
	title: string;
	children: React.ReactNode;
	className?: string;
}) {
	return (
		<section
			className={cn(
				"flex min-w-0 flex-col gap-3 rounded-xl border border-border bg-card p-4",
				className,
			)}
		>
			<h3 className="text-sm font-semibold text-foreground">{title}</h3>
			{children}
		</section>
	);
}

/** RecentActivityRow is one method + URL + status + age line. */
function RecentActivityRow({ entry }: { entry: HistoryEntry }) {
	return (
		<div className="flex min-w-0 items-center gap-2 text-xs">
			<span
				className={cn(
					"font-data shrink-0 rounded-full border border-border bg-muted/40 px-1.5 py-px text-2xs font-semibold uppercase",
					methodTintClass(entry.method),
				)}
			>
				{entry.method}
			</span>
			<span className="min-w-0 flex-1 truncate font-data text-muted-foreground">
				{entry.url}
			</span>
			<StatusPill status={entry.status} />
			<span className="shrink-0 text-muted-foreground/70">
				{relativeTime(entry.createdAt)}
			</span>
		</div>
	);
}

/** OverviewHome is the G-17.3.6 dashboard: workspace hero, quick actions,
 * recent activity, environment summary, repository snapshot, protocol
 * clients. Read-only over existing stores — every action navigates. */
export function OverviewHome() {
	const workspace = useWorkspaceStore((s) => s.currentWorkspace);
	const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const branch = useGitStore((s) => s.status?.branch ?? "");
	const gitFiles = useGitStore((s) => s.status?.files ?? []);
	const repoFound = useGitStore((s) => s.status?.repoFound ?? false);
	const conflicts = useGitStore((s) => s.conflicts);
	const historyPool = useHistoryStore((s) => s.pool);

	const env = environments.find((e) => e.id === activeEnvironmentId);
	const envEntries = Object.entries(env?.variables ?? {}).slice(0, 4);

	const todays = historyPool.filter((e) => isToday(e.createdAt));
	const runsToday = todays.length;
	const passingToday = todays.filter((s) => s.status >= 200 && s.status < 300).length;
	const passingPct =
		runsToday > 0 ? Math.round((passingToday / runsToday) * 100) : 0;
	const avgLatency =
		runsToday > 0
			? `${(todays.reduce((sum, e) => sum + e.durationMs, 0) / runsToday / 1000).toFixed(2)} s`
			: "—";

	const quickActions: QuickAction[] = [
		{
			label: "New request",
			icon: SquareTerminal,
			onSelect: () => {
				openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
				setActiveView("requests");
			},
		},
		{
			label: "Import API",
			icon: Import,
			onSelect: () => setImportOpen(true),
		},
		{
			label: "Spin up a mock",
			icon: Server,
			onSelect: () => setActiveView("mocks"),
		},
		{
			label: "Diff two specs",
			icon: Diff,
			onSelect: () => setActiveView("diff"),
		},
		{
			label: "Generate client",
			icon: FileCode2,
			onSelect: () => setActiveView("docs"),
		},
		{
			label: "Browse OpenAPI",
			icon: FileCode2,
			onSelect: () => setActiveView("explorer"),
		},
	];

	const protocolClients: QuickAction[] = [
		{
			label: "GraphQL",
			icon: SiGraphql,
			onSelect: () => setActiveView("graphql"),
		},
		{
			label: "gRPC",
			icon: Wifi,
			onSelect: () => setActiveView("grpc"),
		},
		{
			label: "WebSocket",
			icon: Wifi,
			onSelect: () => {
				openTab({ id: `realtime-ws-${Date.now()}`, title: "WebSocket", kind: "realtime" });
				setActiveView("requests");
			},
		},
		{
			label: "SSE",
			icon: Play,
			onSelect: () => {
				openTab({ id: `realtime-sse-${Date.now()}`, title: "SSE", kind: "realtime" });
				setActiveView("requests");
			},
		},
	];

	return (
		<ViewShell label="Overview" className="overflow-y-auto">
			<section className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-border bg-card p-4">
				<div className="flex min-w-0 flex-col gap-2">
					<h2 className="text-xl font-semibold text-foreground">
						{workspace?.name ?? "Workspace"}
					</h2>
					<p className="flex items-center gap-1.5 text-xs text-muted-foreground">
						<House className="size-3 shrink-0" aria-hidden />
						Git-native workspace · {workspace?.path ?? "—"}
						{branch ? ` · branch ${branch}` : ""}
					</p>
					<div className="flex flex-wrap items-center gap-1.5">
						<span className="rounded-full border border-primary/25 bg-primary/10 px-2 py-0.5 font-data text-2xs text-primary">
							{countTreeRequests(workspaceTree)} requests
						</span>
						<span className="rounded-full border border-border bg-muted/30 px-2 py-0.5 font-data text-2xs text-muted-foreground">
							{environments.length} environments
						</span>
						<span className="flex items-center gap-1 rounded-full border border-status-ok/25 bg-status-ok/10 px-2 py-0.5 font-data text-2xs text-status-ok">
							<ShieldCheck className="size-3" aria-hidden />
							zero telemetry
						</span>
					</div>
				</div>
				<div className="flex items-center gap-8">
					{[
						{ value: String(runsToday), label: "runs today" },
						{ value: `${passingPct}%`, label: "passing" },
						{ value: avgLatency, label: "avg latency" },
					].map((stat) => (
						<div key={stat.label} className="flex flex-col items-center gap-0.5">
							<span className="font-data text-2xl font-semibold text-primary">
								{stat.value}
							</span>
							<span className="text-xs text-muted-foreground">{stat.label}</span>
						</div>
					))}
				</div>
			</section>

			<div className="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-4">
				<Card title="Quick actions">
					<div className="grid grid-cols-3 gap-2">
						{quickActions.map((a) => (
							<QuickActionTile key={a.label} action={a} />
						))}
					</div>
				</Card>

				<Card title="Recent activity">
					{historyPool.length === 0 ? (
						<Empty className="p-2 text-xs">
							<EmptyHeader>
								<EmptyDescription className="text-xs">
									No requests sent yet — activity lands here.
								</EmptyDescription>
							</EmptyHeader>
						</Empty>
					) : (
						<div className="flex flex-col gap-1.5">
							{historyPool.slice(0, 6).map((e) => (
								<RecentActivityRow key={e.id} entry={e} />
							))}
						</div>
					)}
				</Card>

				<Card title="Environment">
					<p className="text-xs text-muted-foreground">
						Active:{" "}
						<span className="font-medium text-foreground">
							{env?.name ?? "none"}
						</span>
					</p>
					{envEntries.length > 0 ? (
						<div className="flex flex-col gap-1">
							{envEntries.map(([key, value]) => (
								<div
									key={key}
									className="flex items-center justify-between gap-2 font-data text-xs"
								>
									<span className="text-muted-foreground">{`{{${key}}}`}</span>
									<span className="truncate text-foreground">{value}</span>
								</div>
							))}
						</div>
					) : null}
					<button
						type="button"
						onClick={() => setActiveView("environments")}
						className="self-start rounded-md border border-border px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
					>
						Manage environments
					</button>
				</Card>

				<Card title="Repository snapshot">
					{repoFound ? (
						<div className="flex flex-col gap-1.5 text-xs">
							<p className="flex items-center gap-1.5">
								<GitBranch className="size-3 shrink-0" aria-hidden />
								<span className="font-data">{branch || "—"}</span>
							</p>
							<p className="flex items-center gap-1.5">
								<Leaf
									className={cn(
										"size-3 shrink-0",
										gitFiles.length === 0 ? "text-status-ok" : "text-status-warn",
									)}
									aria-hidden
								/>
								{gitFiles.length === 0
									? "working tree clean"
									: `${gitFiles.length} changed file${gitFiles.length === 1 ? "" : "s"}`}
							</p>
							<p className="flex items-center gap-1.5">
								<Diff className="size-3 shrink-0" aria-hidden />
								{conflicts.length === 0
									? "no conflicts"
									: `${conflicts.length} unresolved conflict${conflicts.length === 1 ? "" : "s"}`}
							</p>
						</div>
					) : (
						<p className="text-xs text-muted-foreground">
							Not a git repository — collections still save as plain files.
						</p>
					)}
				</Card>
			</div>

			<Card title="Protocol clients" className="max-w-xl">
				<div className="grid grid-cols-3 gap-2">
					{protocolClients.map((a) => (
						<QuickActionTile key={a.label} action={a} />
					))}
				</div>
			</Card>

			<p className="flex items-center gap-1.5 text-xs text-muted-foreground/60">
				<Clock className="size-3" aria-hidden />
				Stats cover today's sends recorded in local history.
			</p>
		</ViewShell>
	);
}
