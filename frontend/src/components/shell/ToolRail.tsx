import {
	Antenna,
	ArrowLeftRight,
	BookText,
	Cable,
	Compass,
	FileCode2,
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
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "#lib/utils";
import { ThemeToggle } from "../ThemeToggle";
import { Settings } from "lucide-react";
import { useWorkspaceStore, type WorkspaceView } from "../../stores";

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
					"group relative flex items-center justify-center rounded-md transition-colors",
					collapsed ? "size-10" : "h-10 w-full gap-2 px-2",
					active
						? "bg-primary/12 text-primary"
						: "text-muted-foreground hover:bg-muted hover:text-foreground",
				)}
			>
				<span
					aria-hidden
					className={cn(
						"absolute -left-[9px] h-5 w-[3px] rounded-full bg-primary transition-opacity",
						active ? "opacity-100" : "opacity-0",
					)}
				/>
				<Icon className="size-[18px] shrink-0" aria-hidden />
				{!collapsed && (
					<span className="truncate text-sm">{item.label}</span>
				)}
			</button>
		);
	};

	return (
		<nav
			aria-label="Tools"
			className={cn(
				"flex shrink-0 flex-col items-center gap-1 border-r border-border bg-card/40 py-2 transition-all",
				collapsed ? "w-10" : "w-14",
			)}
		>
			<div className="flex flex-col items-center gap-1">
				{WORKSPACE_GROUP.map(railButton)}
			</div>
			<div className="my-1.5 h-px w-6 bg-border" aria-hidden />
			<div className="flex flex-col items-center gap-1">
				{API_TOOLS_GROUP.map(railButton)}
			</div>
			<div className="my-1.5 h-px w-6 bg-border" aria-hidden />
			<div className="flex flex-col items-center gap-1">
				{REALTIME_GROUP.map(railButton)}
			</div>
			<div className="mt-auto flex flex-col items-center gap-1 pt-2">
				{railButton({ view: "settings", label: "Settings", icon: Settings })}
				<ThemeToggle />
				<button
					type="button"
					onClick={onToggleCollapse}
					aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
					title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
					className="flex size-10 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				>
					{collapsed ? (
						<PanelLeftOpen className="size-4" aria-hidden />
					) : (
						<PanelLeftClose className="size-4" aria-hidden />
					)}
				</button>
			</div>
		</nav>
	);
}
