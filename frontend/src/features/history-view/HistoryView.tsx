import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, Download, RotateCcw, Search, Star, Trash2 } from "lucide-react";
import { Button } from "../../components/ui/button";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
} from "../../components/ui/empty";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "../../components/ui/alert-dialog";
import { MethodLabel, StatusPill } from "../../components/status";
import { SplitView, ViewShell } from "../../components/shell/ViewLayout";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "../../components/ui/select";
import { CodeMirrorEditor } from "../../editors";
import { HISTORY_PAGE_SIZE, useHistoryStore } from "../../stores/useHistoryStore";
import { useFuseSearch } from "#hooks/useFuseSearch";
import { useExportStore } from "#stores/useExportStore";
import { HISTORY_FUSE_OPTIONS } from "#lib/historySearch";
import type { HistoryEntry } from "../../lib/history";

const PAGE_SIZE = HISTORY_PAGE_SIZE;
const SAVED_SEARCHES_KEY = "reqly-history-saved-searches.v1";

/** loadSavedSearches reads the persisted saved-search list. Module-level
 * because React Compiler cannot handle try/catch. */
function loadSavedSearches(): string[] {
	try {
		const raw = localStorage.getItem(SAVED_SEARCHES_KEY);
		if (raw == null) return [];
		// SAFETY: localStorage value parsed at the boundary; validated as an
		// array of strings below.
		const parsed = JSON.parse(raw) as unknown;
		return Array.isArray(parsed) ? parsed.filter((s): s is string => typeof s === "string") : [];
	} catch {
		return [];
	}
}

function persistSavedSearches(list: string[]): void {
	try {
		localStorage.setItem(SAVED_SEARCHES_KEY, JSON.stringify(list));
	} catch {
		// storage unavailable — in-memory only
	}
}

/** SavedSearchesPanel is the persisted-search sidebar: save the current
 * query, apply, remove. */
function SavedSearchesPanel({
	currentQuery,
	onApply,
}: {
	currentQuery: string;
	onApply: (query: string) => void;
}) {
	const [list, setList] = useState<string[]>(loadSavedSearches);
	const save = (): void => {
		const q = currentQuery.trim();
		if (q === "" || list.includes(q)) return;
		const next = [q, ...list].slice(0, 10);
		setList(next);
		persistSavedSearches(next);
	};
	const remove = (q: string): void => {
		const next = list.filter((s) => s !== q);
		setList(next);
		persistSavedSearches(next);
	};
	return (
		<div className="flex min-w-0 flex-col gap-1 rounded-xl border border-border bg-card p-3">
			<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
				Saved searches
			</p>
			<button
				type="button"
				onClick={save}
				disabled={currentQuery.trim() === ""}
				className="flex items-center gap-1.5 self-start rounded px-1 py-0.5 text-xs text-muted-foreground transition-colors hover:text-foreground disabled:opacity-40"
			>
				<Star className="size-3.5" aria-hidden />
				Save current search
			</button>
			{list.length === 0 ? (
				<p className="text-xs text-muted-foreground/70">
					Search, then save it to pin it here.
				</p>
			) : (
				list.map((q) => (
					<div key={q} className="flex items-center gap-1">
						<button
							type="button"
							onClick={() => onApply(q)}
							className="min-w-0 flex-1 truncate rounded px-1 py-0.5 text-left text-xs text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
							title={q}
						>
							{q}
						</button>
						<Button
							variant="ghost"
							size="icon-xs"
							aria-label={`Remove saved search ${q}`}
							onClick={() => remove(q)}
						>
							<Trash2 className="size-3.5" />
						</Button>
					</div>
				))
			)}
		</div>
	);
}

function formatTime(iso: string): string {
	if (!iso) return "—";
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return iso;
	return date.toLocaleString();
}

function formatDuration(ms: number): string {
	return ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms}ms`;
}

export function HistoryView() {
	const entries = useHistoryStore((s) => s.entries);
	const loading = useHistoryStore((s) => s.loading);
	const error = useHistoryStore((s) => s.error);
	const pool = useHistoryStore((s) => s.pool);
	const loadPool = useHistoryStore((s) => s.loadPool);
	const replayed = useHistoryStore((s) => s.replayed);
	const load = useHistoryStore((s) => s.load);
	const clear = useHistoryStore((s) => s.clear);
	const replay = useHistoryStore((s) => s.replay);
	const dismissReplay = useHistoryStore((s) => s.dismissReplay);

	const [query, setQuery] = useState("");
	const [status, setStatus] = useState("");
	const [page, setPage] = useState(0);
	const [confirmingClear, setConfirmingClear] = useState(false);

	useEffect(() => {
		void load({ offset: 0, status: "", query: "" });
		if (pool.length === 0) void loadPool();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [load]);

	// Fuse.js fuzzy search over the recent pool — punctuation-safe and
	// typo-tolerant; the plain list renders when the query is blank.
	const searched = useFuseSearch(pool, query, HISTORY_FUSE_OPTIONS);
	const displayEntries = query.trim() === "" ? entries : searched.slice(0, PAGE_SIZE);

	const search = () => {
		setPage(0);
		void load({ offset: 0, status, query });
	};

	const gotoPage = (next: number) => {
		setPage(next);
		void load({ offset: next * PAGE_SIZE, status, query });
	};

	const onReplay = (entry: HistoryEntry) => {
		void replay(entry.id);
	};

	const pageFull = entries.length === PAGE_SIZE;

	const applySavedSearch = (q: string): void => {
		setQuery(q);
		setPage(0);
		void load({ offset: 0, status, query: q });
	};

	const exportHar = (): void => {
		const exportStore = useExportStore.getState();
		exportStore.setFormat("har");
		exportStore.setCollection("");
		void exportStore.run();
	};

	return (
		<ViewShell>
			<SplitView
				asideLabel="Saved searches"
				aside={<SavedSearchesPanel currentQuery={query} onApply={applySavedSearch} />}
			>
			<div className="flex min-w-0 flex-1 flex-col">
			<div className="flex shrink-0 flex-col gap-3 pb-3">
				<div>
					<h2 className="text-lg font-medium">History</h2>
					<p className="text-xs text-muted-foreground">
						Local request history for this workspace — stored on disk, never
						sent anywhere.
					</p>
				</div>
				<div className="flex items-center gap-1">
					<input
						value={query}
						onChange={(e) => setQuery(e.target.value)}
						onKeyDown={(e) => {
							if (e.key === "Enter") search();
						}}
						placeholder="Search URL or path…"
						aria-label="Search history"
						className="min-w-0 flex-1 rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground"
					/>
					<Button size="sm" variant="outline" onClick={search} disabled={loading}>
						<Search className="size-3.5" aria-hidden />
						<span className="sr-only">Search</span>
					</Button>
				</div>
				<div className="flex items-center gap-2">
					<Select
						items={[
							{ value: "", label: "All statuses" },
							{ value: "2xx", label: "2xx" },
							{ value: "4xx", label: "4xx" },
							{ value: "5xx", label: "5xx" },
						]}
						value={status}
						onValueChange={(next) => {
							if (next === null) return;
							setStatus(next);
							setPage(0);
							void load({ offset: 0, status: next, query });
						}}
					>
						<SelectTrigger aria-label="Status filter" className="h-7 w-auto gap-1 rounded-md px-2 text-xs">
							<SelectValue />
						</SelectTrigger>
						<SelectContent className="max-h-72 min-w-(--anchor-width)">
							{[
								{ value: "", label: "All statuses" },
								{ value: "2xx", label: "2xx" },
								{ value: "4xx", label: "4xx" },
								{ value: "5xx", label: "5xx" },
							].map((option) => (
								<SelectItem key={option.value} value={option.value} className="text-xs">
									{option.label}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
					<span className="font-data text-xs tabular-nums text-muted-foreground">
						page {page + 1}
					</span>
					<span className="ml-auto flex items-center gap-1">
						<Button
							size="icon-sm"
							variant="ghost"
							onClick={() => gotoPage(Math.max(0, page - 1))}
							disabled={page === 0 || loading}
							aria-label="Previous page"
						>
							<ChevronLeft className="size-3.5" aria-hidden />
						</Button>
						<Button
							size="icon-sm"
							variant="ghost"
							onClick={() => gotoPage(page + 1)}
							disabled={!pageFull || loading}
							aria-label="Next page"
						>
							<ChevronRight className="size-3.5" aria-hidden />
						</Button>
						<Button
							size="sm"
							variant="outline"
							onClick={exportHar}
							title="Export history to .reqly/exports as HAR 1.2"
						>
							<Download className="size-3.5" aria-hidden />
							HAR
						</Button>
						<Button
							size="sm"
							variant="ghost"
							className="text-destructive hover:bg-destructive/10"
							onClick={() => setConfirmingClear(true)}
						>
							<Trash2 className="size-3.5" aria-hidden />
							Clear
						</Button>
					</span>
				</div>
			</div>

			{error ? (
				<p role="alert" className="shrink-0 text-xs text-destructive">
					{error}
				</p>
			) : null}

			{replayed ? (
				<div className="mb-2 flex shrink-0 flex-col gap-1 rounded-md border border-border bg-muted/30 p-2">
					<div className="flex items-center gap-2">
						<StatusPill status={replayed.statusCode} />
						<span className="font-data text-xs tabular-nums text-muted-foreground">
							replayed · {replayed.durationMs}ms · {replayed.size}B
						</span>
						<Button
							size="sm"
							variant="ghost"
							className="ml-auto"
							onClick={dismissReplay}
						>
							Close
						</Button>
					</div>
					<CodeMirrorEditor
						value={replayed.body}
						language="json"
						readOnly
						className="h-40 overflow-hidden rounded-md border border-border"
					/>
				</div>
			) : null}

			<div className="min-h-0 flex-1 overflow-auto rounded-md border border-border bg-background">
				{displayEntries.length === 0 ? (
					<Empty className="text-xs">
						<EmptyHeader>
							<EmptyDescription>
								{query.trim() === ""
									? "No history yet — send a request and it will be recorded here."
									: "Nothing matches this search."}
							</EmptyDescription>
						</EmptyHeader>
					</Empty>
				) : (
					<table className="w-full border-separate border-spacing-0 text-left text-xs">
						<thead>
							<tr>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">Time</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">Method</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">URL / path</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">Status</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">Duration</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5 font-medium">Env</th>
								<th className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1.5" />
							</tr>
						</thead>
						<tbody>
							{displayEntries.map((entry) => (
								<tr key={entry.id} className="hover:bg-muted/40">
									<td className="border-b border-border/50 px-2 py-1.5 font-data tabular-nums text-muted-foreground">
										{formatTime(entry.createdAt)}
									</td>
									<td className="border-b border-border/50 px-2 py-1.5">
										<MethodLabel method={entry.method} />
									</td>
									<td className="max-w-[280px] border-b border-border/50 px-2 py-1.5">
										<span className="block truncate font-mono" title={entry.url}>
											{entry.url || entry.requestPath || "—"}
										</span>
									</td>
									<td className="border-b border-border/50 px-2 py-1.5">
										<StatusPill status={entry.status} />
									</td>
									<td className="border-b border-border/50 px-2 py-1.5 font-data tabular-nums text-muted-foreground">
										{formatDuration(entry.durationMs)}
									</td>
									<td className="border-b border-border/50 px-2 py-1.5 font-mono text-muted-foreground">
										{entry.env || "—"}
									</td>
									<td className="border-b border-border/50 px-2 py-1.5 text-right">
										<Button
											size="xs"
											variant="ghost"
											onClick={() => onReplay(entry)}
											title={`Replay ${entry.method} ${entry.url}`}
										>
											<RotateCcw className="size-3.5" aria-hidden />
											Replay
										</Button>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				)}
				<p className="shrink-0 px-1 pt-1 font-data text-xs text-muted-foreground/70">
					{displayEntries.length} of {pool.length} entries · stored locally ·
					FTS5 indexed
				</p>
			</div>

			<AlertDialog open={confirmingClear} onOpenChange={setConfirmingClear}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Clear all history?</AlertDialogTitle>
						<AlertDialogDescription>
							This removes every stored request/response record for this
							workspace from local disk. It cannot be undone.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction
							className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
							onClick={() => {
								setPage(0);
								void clear(null);
							}}
						>
							Clear history
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
			</div>
			</SplitView>
		</ViewShell>
	);
}
