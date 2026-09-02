import { useEffect, useState } from "react";
import { Copy, Pin, X, Plus } from "lucide-react";
import { useRealtimeStore } from "#stores/useRealtimeStore";
import { useTestStore } from "#stores/useTestStore";
import { useWorkspaceStore } from "#stores";
import { NEW_REQUEST_TAB_ID, tabIsDirty } from "#stores/useRequestStore";
import { useRequestStore } from "#stores/useRequestStore";
import type { RequestTab } from "#stores/useWorkspaceStore";
import { cn } from "#lib/utils";
import { handleTabArrowKeys } from "#lib/ui";
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

function TabItem({ tab, index, onReorder }: { tab: RequestTab; index: number; onReorder: (from: number, to: number) => void }) {
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const setActiveTab = useWorkspaceStore((s) => s.setActiveTab);
	const closeTab = useWorkspaceStore((s) => s.closeTab);
	const dirty = useRequestStore((s) =>
		tabIsDirty(s.drafts[tab.id], s.meta[tab.id]),
	);
	const [confirming, setConfirming] = useState(false);
	const [pinned, setPinned] = useState(false);
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

	const duplicate = () => {
		if ((tab.kind ?? "request") === "request") duplicateTab(tab.id);
	};

	return (
		<div
			onContextMenu={(e) => {
				e.preventDefault();
				setMenu({ x: e.clientX, y: e.clientY });
			}}
			draggable
			onDragStart={(e) => e.dataTransfer.setData("text/plain", String(index))}
			onDragOver={(e) => e.preventDefault()}
			onDrop={(e) => {
				e.preventDefault();
				const from = Number(e.dataTransfer.getData("text/plain"));
				if (!Number.isNaN(from) && from !== index) onReorder(from, index);
			}}
			className={cn(
				"group relative flex h-8 shrink-0 items-center gap-1 border-b-2 px-2.5 text-xs transition-colors select-none",
				active
					? "border-primary bg-card text-foreground"
					: "border-transparent bg-transparent text-muted-foreground hover:bg-muted/40 hover:text-foreground",
			)}
		>
			<button
				type="button"
				role="tab"
				aria-selected={active}
				tabIndex={active ? 0 : -1}
				onClick={() => setActiveTab(tab.id)}
				className="flex max-w-40 items-center truncate text-left font-mono text-xs"
				title={tab.title}
			>
				{pinned && <Pin className="mr-1 size-2.5 shrink-0 text-muted-foreground" aria-label="Pinned" />}
				{dirty && (
					<span
						title="Unsaved changes"
						aria-label="Unsaved changes"
						className="mr-1.5 inline-block size-1.5 rounded-full bg-status-warn"
					/>
				)}
				<span className="truncate">{tab.title}</span>
			</button>
			<div className="ml-1 hidden items-center gap-0.5 group-hover:flex sm:flex">
				<button
					type="button"
					onClick={() => setPinned((v) => !v)}
					title={pinned ? "Unpin" : "Pin"}
					aria-label={pinned ? "Unpin tab" : "Pin tab"}
					className={cn(
						"rounded p-1 transition-colors hover:bg-muted",
						pinned ? "text-foreground opacity-100" : "text-muted-foreground/60 opacity-60 hover:text-foreground hover:opacity-100",
					)}
				>
					<Pin className="size-3" aria-hidden />
				</button>
				<button
					type="button"
					onClick={duplicate}
					title="Duplicate tab"
					aria-label={`Duplicate ${tab.title}`}
					className="rounded p-1 text-muted-foreground/60 opacity-60 transition-colors hover:bg-muted hover:text-foreground hover:opacity-100"
				>
					<Copy className="size-3" aria-hidden />
				</button>
				<button
					type="button"
					onClick={requestClose}
					title="Close tab"
					aria-label={`Close ${tab.title}`}
					className="rounded p-1 text-muted-foreground/60 opacity-60 transition-colors hover:bg-muted hover:text-destructive hover:opacity-100"
				>
					<X className="size-3" aria-hidden />
				</button>
			</div>
			{/* mobile close always visible when active */}
			<button
				type="button"
				onClick={requestClose}
				className={cn(
					"ml-1 rounded p-1 sm:hidden",
					active ? "text-muted-foreground hover:text-destructive" : "hidden",
				)}
				aria-label={`Close ${tab.title}`}
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
					]}
					onClose={() => setMenu(null)}
				/>
			)}
		</div>
	);
}

/**
 * Request tab bar — one tab per open request, quiet underline indicator,
 * no double chrome. Overlap with editor border removed.
 */
export function RequestTabs() {
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setTabs = (tabs: typeof openTabs) => useWorkspaceStore.setState({ openTabs: tabs });
	const reorder = (from: number, to: number) => {
		const next = [...openTabs];
		const [moved] = next.splice(from, 1);
		next.splice(to, 0, moved);
		setTabs(next);
	};

	useEffect(() => {
		try {
			const raw = localStorage.getItem("reqly:tabs:v1") ?? localStorage.getItem("reqly:tabs");
			if (raw) {
				// SAFETY: raw is JSON from localStorage written by this app; parsed shape validated by Array.isArray check below.
				const parsed = JSON.parse(raw) as { openTabs: typeof openTabs; activeTabId: string | null };
				if (Array.isArray(parsed.openTabs) && parsed.openTabs.length) {
					useWorkspaceStore.setState({ openTabs: parsed.openTabs, activeTabId: parsed.activeTabId });
				}
			}
		} catch {}
	}, []);
	useEffect(() => {
		const activeTabId = useWorkspaceStore.getState().activeTabId;
		localStorage.setItem("reqly:tabs:v1", JSON.stringify({ openTabs, activeTabId }));
	}, [openTabs]);

	return (
		<div
			role="tablist"
			aria-label="Open requests"
			onKeyDown={(e) => handleTabArrowKeys(e)}
			className="flex h-8 shrink-0 items-center gap-0 overflow-x-auto border-b border-border bg-card px-1"
		>
			{openTabs.map((t, i) => (
				<TabItem key={t.id} tab={t} index={i} onReorder={reorder} />
			))}
			<button
				type="button"
				onClick={() =>
					openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" })
				}
				title="New request"
				className="ml-1 shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
			>
				<Plus className="size-3.5" aria-hidden />
				<span className="sr-only">New request</span>
			</button>
		</div>
	);
}
