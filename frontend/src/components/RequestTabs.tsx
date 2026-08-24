import { useState } from "react";
import { X, Plus } from "lucide-react";
import { useTestStore } from "#stores/useTestStore";
import { useWorkspaceStore } from "#stores";
import { NEW_REQUEST_TAB_ID, tabIsDirty } from "#stores/useRequestStore";
import { useRequestStore } from "#stores/useRequestStore";
import type { RequestTab } from "#stores/useWorkspaceStore";
import { cn } from "#lib/utils";
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

function TabItem({ tab }: { tab: RequestTab }) {
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const setActiveTab = useWorkspaceStore((s) => s.setActiveTab);
	const closeTab = useWorkspaceStore((s) => s.closeTab);
	const dirty = useRequestStore((s) =>
		tabIsDirty(s.drafts[tab.id], s.meta[tab.id]),
	);
	const [confirming, setConfirming] = useState(false);
	const active = activeTabId === tab.id;

	const requestClose = () => {
		if (dirty) {
			setConfirming(true);
			return;
		}
		if (tab.kind === "test") useTestStore.getState().closeTab(tab.id);
		closeTab(tab.id, { force: true });
	};

	return (
		<div
			className={cn(
				"group flex shrink-0 items-center gap-1 rounded-md border px-2 py-1 text-xs transition-colors",
				active
					? "border-border bg-muted text-foreground"
					: "border-transparent text-muted-foreground hover:bg-muted/50 hover:text-foreground",
			)}
		>
			<button
				type="button"
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
			<button
				type="button"
				onClick={requestClose}
				onAuxClick={(e) => {
					if (e.button === 1) requestClose();
				}}
				title="Close tab (middle-click)"
				aria-label={`Close ${tab.title}`}
				className="rounded p-0.5 text-muted-foreground/50 opacity-0 transition-opacity focus-visible:opacity-100 group-hover:opacity-100 hover:text-foreground"
			>
				<X className="size-3" aria-hidden />
			</button>
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
			className="flex shrink-0 items-center gap-1 overflow-x-auto border-b border-border px-2 py-1"
		>
			{openTabs.map((t) => (
				<TabItem key={t.id} tab={t} />
			))}
			<button
				type="button"
				onClick={() =>
					openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" })
				}
				title="New request"
				className="shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
			>
				<Plus className="size-3.5" aria-hidden />
				<span className="sr-only">New request</span>
			</button>
		</div>
	);
}
