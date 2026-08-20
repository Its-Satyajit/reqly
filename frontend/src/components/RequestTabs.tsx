import { cn } from "#lib/utils";
import { useWorkspaceStore } from "#stores";
import { NEW_REQUEST_TAB_ID } from "#stores/useRequestStore";

/**
 * The request tab bar: one tab per open request (deduplicated by id) plus a
 * "+ New request" action that focuses the persistent scratchpad tab.
 */
export function RequestTabs() {
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const setActiveTab = useWorkspaceStore((s) => s.setActiveTab);
	const closeTab = useWorkspaceStore((s) => s.closeTab);
	const openTab = useWorkspaceStore((s) => s.openTab);

	const newRequest = () => {
		openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
	};

	return (
		<div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b border-border px-2 py-1">
			{openTabs.map((t) => (
				<div
					key={t.id}
					className={cn(
						"group flex shrink-0 items-center gap-1 rounded-md border border-border px-2 py-1 text-xs transition-colors",
						activeTabId === t.id
							? "bg-muted text-foreground"
							: "text-muted-foreground hover:text-foreground",
					)}
				>
					<button
						onClick={() => setActiveTab(t.id)}
						className="max-w-40 truncate"
					>
						{t.title}
					</button>
					<button
						onClick={() => closeTab(t.id)}
						title="Close tab"
						className="rounded p-0.5 text-muted-foreground/60 transition-colors hover:bg-background hover:text-foreground"
					>
						<svg
							className="size-3"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							strokeWidth="2"
							strokeLinecap="round"
							strokeLinejoin="round"
							aria-hidden
						>
							<path d="M18 6 6 18M6 6l12 12" />
						</svg>
					</button>
				</div>
			))}
			<button
				onClick={newRequest}
				className="shrink-0 rounded-md px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
			>
				+ New request
			</button>
		</div>
	);
}
