import { useState } from "react";
import { X, Plus } from "lucide-react";
import { useRealtimeStore } from "#stores/useRealtimeStore";
import { useTestStore } from "#stores/useTestStore";
import { useWorkspaceStore } from "#stores";
import { NEW_REQUEST_TAB_ID, tabIsDirty } from "#stores/useRequestStore";
import { useRequestStore } from "#stores/useRequestStore";
import type { RequestTab } from "#stores/useWorkspaceStore";
import { cn } from "#lib/utils";
import { handleTabArrowKeys } from "#lib/ui";
import { Button } from "#components/ui/button";
import { ContextMenu } from "#components/ContextMenu";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#components/ui/alert-dialog";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#components/ui/tooltip";

function TabItem({ tab }: { tab: RequestTab }) {
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const setActiveTab = useWorkspaceStore((s) => s.setActiveTab);
	const closeTab = useWorkspaceStore((s) => s.closeTab);
	const dirty = useRequestStore((s) =>
		tabIsDirty(s.drafts[tab.id], s.meta[tab.id]),
	);
	const [confirming, setConfirming] = useState(false);
	const [menu, setMenu] = useState<{ x: number; y: number } | null>(null);
	const duplicateTab = useWorkspaceStore((s) => s.duplicateTab);
	const active = activeTabId === tab.id;

	const requestClose = () => {
		if (dirty) {
			setConfirming(true);
			return;
		}
		if (tab.kind === "test") useTestStore.getState().closeTab(tab.id);
		if (tab.kind === "realtime") useRealtimeStore.getState().closeTab(tab.id);
		closeTab(tab.id, { force: true });
	};

	return (
		<div
			onContextMenu={(e) => {
				e.preventDefault();
				setMenu({ x: e.clientX, y: e.clientY });
			}}
			className={cn(
				"group flex shrink-0 items-center gap-1 rounded-md border px-2 py-1 text-xs transition-colors",
				active
					? "border-border bg-muted text-foreground"
					: "border-transparent text-muted-foreground hover:bg-muted/50 hover:text-foreground",
			)}
		>
			<button
				type="button"
				role="tab"
				aria-selected={active}
				tabIndex={active ? 0 : -1}
				onClick={() => setActiveTab(tab.id)}
				className="max-w-40 truncate"
				title={tab.title}
			>
				{dirty && (
					<span
						title="Unsaved changes"
						aria-label="Unsaved changes"
						className="mr-1 inline-block size-1.5 rounded-full bg-warning"
					/>
				)}
				{tab.title}
			</button>
			<Tooltip>
				<TooltipTrigger
					render={
						<Button
							type="button"
							variant="ghost"
							size="icon-xs"
							onClick={requestClose}
							onAuxClick={(e) => {
								if (e.button === 1) requestClose();
							}}
							aria-label={`Close ${tab.title}`}
							className="size-5 rounded-sm p-0 text-muted-foreground/50 opacity-0 transition-opacity focus-visible:opacity-100 group-hover:opacity-100 hover:text-foreground"
						/>
					}
				>
					<X className="size-3.5" aria-hidden />
				</TooltipTrigger>
				<TooltipContent>Close tab (middle-click)</TooltipContent>
			</Tooltip>
			<AlertDialog open={confirming} onOpenChange={setConfirming}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Discard unsaved changes?</AlertDialogTitle>
						<AlertDialogDescription>
							{tab.title} has changes that were never saved to disk. Closing
							the tab discards them.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Keep editing</AlertDialogCancel>
						<AlertDialogAction
							onClick={() => closeTab(tab.id, { force: true })}
						>
							Discard and close
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
			{menu && (
				<ContextMenu
					x={menu.x}
					y={menu.y}
					items={[
						{
							label: "Duplicate request",
							onSelect: () => {
							if ((tab.kind ?? "request") === "request") duplicateTab(tab.id);
						},
						},
						{ label: "Close tab", onSelect: requestClose },
						{
							label: "Close other tabs",
							onSelect: () => {
								const store = useWorkspaceStore.getState();
								for (const t of store.openTabs) {
									if (t.id !== tab.id) store.closeTab(t.id, { force: true });
								}
							},
						},
					]}
					onClose={() => setMenu(null)}
				/>
			)}
		</div>
	);
}

/**
 * The request tab bar: one tab per open request (deduplicated by id) plus a
 * "+ New request" action that focuses the persistent scratchpad tab. A dot
 * marks file-backed tabs with unsaved edits.
 */
export function RequestTabs() {
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const openTab = useWorkspaceStore((s) => s.openTab);

	return (
		<div
			role="tablist"
			aria-label="Open requests"
			onKeyDown={(e) => handleTabArrowKeys(e)}
			className="flex shrink-0 items-center gap-1 overflow-x-auto border-b border-border px-2 py-1"
		>
			{openTabs.map((t) => (
				<TabItem key={t.id} tab={t} />
			))}
			<Button
				type="button"
				variant="ghost"
				size="icon-sm"
				className="shrink-0 text-muted-foreground"
				onClick={() =>
					openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" })
				}
				title="New request"
			>
				<Plus className="size-3.5" aria-hidden />
				<span className="sr-only">New request</span>
			</Button>
		</div>
	);
}
