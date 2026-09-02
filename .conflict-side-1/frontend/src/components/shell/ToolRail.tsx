import {
	Antenna,
	ArrowLeftRight,
	BookText,
	Cable,
	Compass,
	FileCode2,
	FileDiff,
	Hexagon,
	History,
	House,
	KeyRound,
	PanelLeftClose,
	PanelLeftOpen,
	Play,
	Radio,
	Rss,
	Zap,
	Database,
	Shield,
	ScrollText,
	Fingerprint,
	Users,
	Clock,
	Workflow,
	Activity,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "#lib/utils";
import { ThemeToggle } from "../ThemeToggle";
import { useWorkspaceStore, type WorkspaceView } from "../../stores/useWorkspaceStore";

interface RailItem {
	view?: WorkspaceView;
	label: string;
	icon: LucideIcon;
	action?: () => void;
}

const WORKSPACE_GROUP: RailItem[] = [
	{ view: "home", label: "Workspace", icon: House },
	{ view: "requests", label: "Requests", icon: Zap },
	{ view: "environments", label: "Environments", icon: Database },
	{ view: "history", label: "History", icon: History },
];

const API_TOOLS_GROUP: RailItem[] = [
	{ view: "mocks", label: "Mocks", icon: Antenna },
	{ view: "diff", label: "Diff", icon: ArrowLeftRight },
	{ view: "jwt", label: "JWT", icon: KeyRound },
	{ view: "graphql", label: "GraphQL", icon: Hexagon },
	{ view: "grpc", label: "gRPC", icon: Cable },
	{ view: "runners", label: "Runners", icon: Play },
	{ view: "explorer", label: "Explorer", icon: Compass },
	{ view: "docs", label: "Docs", icon: BookText },
	{ view: "spec-editor", label: "Spec Editor", icon: FileCode2 },
];

const REALTIME_GROUP: RailItem[] = [
	// SAFETY: "websocket" is a valid WorkspaceView value
	{ view: "websocket" as WorkspaceView, label: "WebSocket", icon: Radio },
	// SAFETY: "sse" is a valid WorkspaceView value
	{ view: "sse" as WorkspaceView, label: "SSE", icon: Rss },
	// SAFETY: "mqtt" is a valid WorkspaceView value
	{ view: "mqtt" as WorkspaceView, label: "MQTT", icon: Antenna },
	// SAFETY: "socketio" is a valid WorkspaceView value
	{ view: "socketio" as WorkspaceView, label: "Socket.IO", icon: Cable },
];

const GOVERNANCE_GROUP: RailItem[] = [
	// SAFETY: "policy" is a valid WorkspaceView value
	{ view: "policy" as WorkspaceView, label: "Policy & RBAC", icon: Shield },
	// SAFETY: "audit" is a valid WorkspaceView value
	{ view: "audit" as WorkspaceView, label: "Audit Log", icon: ScrollText },
	// SAFETY: "sso" is a valid WorkspaceView value
	{ view: "sso" as WorkspaceView, label: "SSO & SCIM", icon: Fingerprint },
	// SAFETY: "collab" is a valid WorkspaceView value
	{ view: "collab" as WorkspaceView, label: "Collaboration", icon: Users },
];

const ORCHESTRATION_GROUP: RailItem[] = [
	// SAFETY: "automation" is a valid WorkspaceView value
	{ view: "automation" as WorkspaceView, label: "Automation", icon: Clock },
	// SAFETY: "workflow" is a valid WorkspaceView value
	{ view: "workflow" as WorkspaceView, label: "Workflow", icon: Workflow },
	// SAFETY: "monitor" is a valid WorkspaceView value
	{ view: "monitor" as WorkspaceView, label: "Monitor", icon: Activity },
	// SAFETY: "changelog" is a valid WorkspaceView value
	{ view: "changelog" as WorkspaceView, label: "Changelog", icon: FileDiff },
];

interface ToolRailProps {
	collapsed: boolean;
	onToggleCollapse: () => void;
}

export function ToolRail({ collapsed, onToggleCollapse }: ToolRailProps) {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const requestView = useWorkspaceStore((s) => s.requestView);

	const railButton = (item: RailItem) => {
		const active = item.view != null && activeView === item.view;
		const Icon = item.icon;
		return (
			<button
				key={item.label}
				type="button"
				onClick={() =>
					item.action ? item.action() : item.view && requestView(item.view)
				}
				aria-current={active ? "page" : undefined}
				aria-label={item.label}
				title={item.label}
				className={cn(
					"group relative flex items-center rounded transition-all select-none",
					collapsed ? "size-8 justify-center" : "h-8 w-full gap-2.5 px-2.5",
					active
						? "bg-primary/10 font-medium text-primary"
						: "text-muted-foreground hover:bg-muted/70 hover:text-foreground",
				)}
			>
				<span
					aria-hidden
					className={cn(
						"absolute left-0 top-1.5 bottom-1.5 w-[2.5px] rounded-r bg-primary transition-opacity",
						active ? "opacity-100" : "opacity-0",
					)}
				/>
				<Icon className="size-4 shrink-0" aria-hidden />
				{!collapsed && (
					<span className="truncate text-xs tracking-tight">{item.label}</span>
				)}
			</button>
		);
	};

	return (
		<aside
			aria-label="Tool rail"
			className={cn(
				"flex shrink-0 flex-col justify-between border-r border-border bg-card/25 p-1.5 transition-[width] duration-150 ease-out",
				collapsed ? "w-11" : "w-44",
			)}
		>
			<div className="flex flex-col gap-3">
				<div className="flex flex-col gap-0.5">
					{!collapsed && (
						<p className="px-2 pb-1 text-[10px] font-mono font-semibold uppercase tracking-wider text-muted-foreground/70">
							Workspace
						</p>
					)}
					{WORKSPACE_GROUP.map(railButton)}
				</div>

				<div className="flex flex-col gap-0.5 border-t border-border/60 pt-2">
					{!collapsed && (
						<p className="px-2 pb-1 text-[10px] font-mono font-semibold uppercase tracking-wider text-muted-foreground/70">
							API Tools
						</p>
					)}
					{API_TOOLS_GROUP.map(railButton)}
				</div>

				<div className="flex flex-col gap-0.5 border-t border-border/60 pt-2">
					{!collapsed && (
						<p className="px-2 pb-1 text-[10px] font-mono font-semibold uppercase tracking-wider text-muted-foreground/70">
							Realtime
						</p>
					)}
					{REALTIME_GROUP.map(railButton)}
				</div>

				<div className="flex flex-col gap-0.5 border-t border-border/60 pt-2">
					{!collapsed && (
						<p className="px-2 pb-1 text-[10px] font-mono font-semibold uppercase tracking-wider text-muted-foreground/70">
							Governance
						</p>
					)}
					{GOVERNANCE_GROUP.map(railButton)}
				</div>

				<div className="flex flex-col gap-0.5 border-t border-border/60 pt-2">
					{!collapsed && (
						<p className="px-2 pb-1 text-[10px] font-mono font-semibold uppercase tracking-wider text-muted-foreground/70">
							Orchestration
						</p>
					)}
					{ORCHESTRATION_GROUP.map(railButton)}
				</div>
			</div>

			<div className="flex flex-col gap-1 border-t border-border/60 pt-2">
				<div
					className={cn(
						"flex items-center",
						collapsed ? "flex-col gap-1" : "justify-between px-1",
					)}
				>
					<ThemeToggle />
					<button
						type="button"
						onClick={onToggleCollapse}
						aria-label={collapsed ? "Expand tool rail" : "Collapse tool rail"}
						title={collapsed ? "Expand rail" : "Collapse rail"}
						className="flex size-7 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					>
						{collapsed ? (
							<PanelLeftOpen className="size-4" aria-hidden />
						) : (
							<PanelLeftClose className="size-4" aria-hidden />
						)}
					</button>
				</div>
			</div>
		</aside>
	);
}
