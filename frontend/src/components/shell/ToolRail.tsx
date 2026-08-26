import {
	Antenna,
	ArrowLeftRight,
	BookText,
	Cable,
	Compass,
	Hexagon,
	History,
	House,
	KeyRound,
	Play,
	Radio,
	Rss,
	Zap,
	Database,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "#lib/utils";
import { ThemeToggle } from "../ThemeToggle";
import {
	useRealtimeStore,
	useWorkspaceStore,
	type WorkspaceView,
} from "../../stores";

interface RailItem {
	view?: WorkspaceView;
	label: string;
	icon: LucideIcon;
	action?: () => void;
}

const WORKSPACE_GROUP: RailItem[] = [
	{ view: "home", label: "Workspace home", icon: House },
	{ view: "requests", label: "Requests", icon: Zap },
	{ view: "environments", label: "Environments", icon: Database },
	{ view: "history", label: "History", icon: History },
];

const API_TOOLS_GROUP: RailItem[] = [
	{ view: "mocks", label: "Mocks", icon: Antenna },
	{ view: "diff", label: "Diff", icon: ArrowLeftRight },
	{ view: "jwt", label: "JWT Inspector", icon: KeyRound },
	{ view: "graphql", label: "GraphQL", icon: Hexagon },
	{ view: "grpc", label: "gRPC", icon: Cable },
	{ view: "runners", label: "Runners", icon: Play },
	{ view: "explorer", label: "Explorer", icon: Compass },
	{ view: "docs", label: "Docs", icon: BookText },
];

export function ToolRail({ className }: { className?: string }) {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const requestView = useWorkspaceStore((s) => s.requestView);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const newRealtimeTab = useRealtimeStore((s) => s.newTab);

	const realtimeGroup: RailItem[] = [
		{
			label: "WebSocket",
			icon: Radio,
			action: () => {
				const id = `realtime-ws-${Date.now()}`;
				newRealtimeTab(id, "ws");
				openTab({ id, title: "WebSocket", kind: "realtime" });
				requestView("requests");
			},
		},
		{
			label: "Server-sent events",
			icon: Rss,
			action: () => {
				const id = `realtime-sse-${Date.now()}`;
				newRealtimeTab(id, "sse");
				openTab({ id, title: "SSE", kind: "realtime" });
				requestView("requests");
			},
		},
	];

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
					"group relative flex size-10 items-center justify-center rounded-md transition-colors",
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
				<Icon className="size-[18px]" aria-hidden />
			</button>
		);
	};

	return (
		<nav
			aria-label="Tools"
			className={cn(
				"flex w-[52px] shrink-0 flex-col items-center gap-1 border-r border-border bg-card/40 py-2",
				className,
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
				{realtimeGroup.map(railButton)}
			</div>
			<div className="mt-auto flex flex-col items-center gap-1 pt-2">
				<ThemeToggle />
			</div>
		</nav>
	);
}
