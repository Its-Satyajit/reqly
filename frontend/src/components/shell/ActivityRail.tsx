import type { ComponentType } from "react";
import {
	BookOpen,
	Clock,
	FileCode2,
	GitBranch,
	GitCompare,
	House,
	KeyRound,
	ListChecks,
	Network,
	Play,
	Radio,
	Server,
	Settings,
	SlidersHorizontal,
	SquareTerminal,
	type LucideIcon,
} from "lucide-react";
import { SiGraphql } from "@icons-pack/react-simple-icons";

import { cn } from "#lib/utils";
import { notifyWarning } from "#lib/notify";
import { useWorkspaceStore, type WorkspaceView } from "#stores";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../ui/tooltip";

interface RailItem {
	view: WorkspaceView;
	label: string;
	/** lucide-react for app icons; @icons-pack/react-simple-icons for brand
	 * icons (protocols with an official mark). Both accept className. */
	icon: ComponentType<{ className?: string }> | LucideIcon;
}

interface RailGroup {
	items: RailItem[];
}

const RAIL_GROUPS: RailGroup[] = [
	{
		items: [{ view: "overview", label: "Overview", icon: House }],
	},
	{
		items: [
			{ view: "requests", label: "REST client", icon: SquareTerminal },
			{ view: "graphql", label: "GraphQL", icon: SiGraphql },
			{ view: "grpc", label: "gRPC", icon: Network },
			{ view: "realtime", label: "Realtime", icon: Radio },
		],
	},
	{
		items: [
			{ view: "tests", label: "Tests", icon: ListChecks },
			{ view: "git", label: "Git", icon: GitBranch },
			{ view: "oauth", label: "OAuth tokens", icon: KeyRound },
			{ view: "environments", label: "Environments", icon: SlidersHorizontal },
			{ view: "history", label: "History", icon: Clock },
			{ view: "runners", label: "Runners", icon: Play },
			{ view: "explorer", label: "OpenAPI explorer", icon: FileCode2 },
			{ view: "mocks", label: "Mock servers", icon: Server },
			{ view: "diff", label: "API diff", icon: GitCompare },
			{ view: "jwt", label: "JWT inspector", icon: KeyRound },
			{ view: "docs", label: "Docs generator", icon: BookOpen },
		],
	},
];

const SETTINGS_ITEM: RailItem = {
	view: "settings",
	label: "Settings",
	icon: Settings,
};

/**
 * Atlas-style icon activity rail (G-17.3.1): views as icon buttons on the
 * left edge, Settings pinned at the bottom. Collapses to a 48px icon strip;
 * expands with labels on hover/focus like the design-9 reference rail.
 */
export function ActivityRail() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const hasUnsavedEnvChanges = useWorkspaceStore(
		(s) => s.hasUnsavedEnvChanges,
	);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);

	const select = (view: WorkspaceView) => {
		if (view === activeView) return;
		if (hasUnsavedEnvChanges) {
			notifyWarning(
				"Environment has unsaved changes",
				"Save or discard them before switching views.",
			);
			return;
		}
		setActiveView(view);
	};

	const railButton = (item: RailItem) => {
		const Icon = item.icon;
		const active = activeView === item.view;
		return (
			<Tooltip key={item.view}>
				<TooltipTrigger
					render={
						<button
							type="button"
							onClick={() => select(item.view)}
							aria-label={item.label}
							aria-current={active ? "page" : undefined}
							className={cn(
								"flex h-9 w-full shrink-0 items-center gap-2.5 overflow-hidden rounded-xl px-3 text-left text-xs whitespace-nowrap transition-colors",
								active
									? "bg-[var(--sel)] font-medium text-primary"
									: "text-muted-foreground hover:bg-accent hover:text-foreground",
							)}
						>
							<Icon className="size-4 shrink-0" aria-hidden />
							<span className="max-w-0 opacity-0 transition-[max-width,opacity] duration-200 group-hover/rail:max-w-32 group-hover/rail:opacity-100 group-focus-within/rail:max-w-32 group-focus-within/rail:opacity-100">
								{item.label}
							</span>
						</button>
					}
				/>
				<TooltipContent side="right">{item.label}</TooltipContent>
			</Tooltip>
		);
	};

	return (
		<nav
			aria-label="Views"
			className="group/rail flex w-12 shrink-0 flex-col gap-2 overflow-hidden border-r border-border bg-background py-2 transition-[width] duration-200 hover:w-48 focus-within:w-48"
		>
			{RAIL_GROUPS.map((group, groupIndex) => (
				<div
					key={group.items[0].view}
					className={cn(
						"flex flex-col gap-0.5 px-2",
						groupIndex > 0 && "border-t border-border pt-2",
					)}
				>
					{group.items.map(railButton)}
				</div>
			))}
			<div className="mt-auto flex flex-col gap-0.5 border-t border-border px-2 pt-2">
				{railButton(SETTINGS_ITEM)}
			</div>
		</nav>
	);
}
