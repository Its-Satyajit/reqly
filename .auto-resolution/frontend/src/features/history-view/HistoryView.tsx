import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, RotateCcw, Search, Trash2 } from "lucide-react";
import { Button } from "../../components/ui/button";
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
import { CompactSelect } from "../../components/CompactSelect";
import { CodeMirrorEditor } from "../../editors";
import { KeyValueEditor } from "../../components/KeyValueEditor";
import type { KeyValueRow } from "../../lib/request";
import { HISTORY_PAGE_SIZE, useHistoryStore } from "../../stores/useHistoryStore";
import { useFuseSearch } from "#hooks/useFuseSearch";
import { HISTORY_FUSE_OPTIONS } from "#lib/historySearch";
import type { HistoryEntry } from "../../lib/history";

const PAGE_SIZE = HISTORY_PAGE_SIZE;

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
	const [methodFilter, setMethodFilter] = useState("");
	const [page, setPage] = useState(0);
	const [confirmingClear, setConfirmingClear] = useState(false);
	const [replayVarsId, setReplayVarsId] = useState<string | null>(null);
	const [replayVars, setReplayVars] = useState<KeyValueRow[]>([{ key: "", value: "", enabled: true }]);

	useEffect(() => {
		void load({ offset: 0, status: "", query: "" });
		if (pool.length === 0) void loadPool();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [load]);

	// Fuse.js fuzzy search over the recent pool
	const searched = useFuseSearch(pool, query, HISTORY_FUSE_OPTIONS);
	const rawDisplay = query.trim() === "" ? entries : searched.slice(0, PAGE_SIZE);

	const displayEntries = rawDisplay.filter((entry) => {
		if (methodFilter && entry.method.toUpperCase() !== methodFilter.toUpperCase()) return false;
		if (status) {
			const s = entry.status;
			if (status === "2xx" && (s < 200 || s >= 300)) return false;
			if (status === "3xx" && (s < 300 || s >= 400)) return false;
			if (status === "4xx" && (s < 400 || s >= 500)) return false;
			if (status === "5xx" && s < 500) return false;
		}
		return true;
	});

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
	const onReplayWithVars = () => {
		if (!replayVarsId) return;
		const vars: Record<string, string> = {};
		for (const r of replayVars) {
			if (r.enabled && r.key.trim() !== "") vars[r.key] = r.value;
		}
		void useHistoryStore.getState().replayWithVars(replayVarsId, vars);
		setReplayVarsId(null);
	};

	const pageFull = entries.length === PAGE_SIZE;

	return (
		<div className="flex h-full min-h-0 flex-col p-4 bg-background">
			<div className="flex shrink-0 flex-col gap-2.5 pb-3">
				<div>
					<h2 className="text-sm font-semibold">History</h2>
					<p className="text-xs text-muted-foreground">
						Local request history for this workspace — stored on disk, never sent anywhere.
					</p>
				</div>
				<div className="flex items-center gap-1.5">
					<div className="flex flex-1 items-center rounded border border-input bg-background focus-within:border-ring">
						<input
							value={query}
							onChange={(e) => setQuery(e.target.value)}
							onKeyDown={(e) => {
								if (e.key === "Enter") search();
							}}
							placeholder="Filter by URL or path…"
							aria-label="Search history"
							className="min-w-0 flex-1 border-0 bg-transparent px-2.5 py-1 font-mono text-xs text-foreground placeholder:text-muted-foreground/60 focus:outline-none"
						/>
						<Button size="xs" variant="ghost" onClick={search} disabled={loading} className="mr-1 h-6 px-2">
							<Search className="size-3" aria-hidden />
						</Button>
					</div>
					<CompactSelect
						value={methodFilter}
						onChange={(m) => setMethodFilter(m)}
						ariaLabel="Method filter"
						className="h-7 w-24 font-mono text-[11px]"
						options={[
							{ value: "", label: "All methods" },
							{ value: "GET", label: "GET" },
							{ value: "POST", label: "POST" },
							{ value: "PUT", label: "PUT" },
							{ value: "PATCH", label: "PATCH" },
							{ value: "DELETE", label: "DELETE" },
						]}
					/>
					<CompactSelect
						value={status}
						onChange={(next) => {
							setStatus(next);
							setPage(0);
							void load({ offset: 0, status: next, query });
						}}
						ariaLabel="Status filter"
						className="h-7 w-28 font-mono text-[11px]"
						options={[
							{ value: "", label: "All statuses" },
							{ value: "2xx", label: "2xx Success" },
							{ value: "3xx", label: "3xx Redirect" },
							{ value: "4xx", label: "4xx Client Err" },
							{ value: "5xx", label: "5xx Server Err" },
						]}
					/>
					<span className="ml-auto flex items-center gap-1">
						<span className="font-mono text-xs tabular-nums text-muted-foreground mr-1">
							page {page + 1}
						</span>
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
							size="xs"
							variant="ghost"
							className="h-7 text-destructive hover:bg-destructive/10 font-mono text-[11px] gap-1"
							onClick={() => setConfirmingClear(true)}
						>
							<Trash2 className="size-3" aria-hidden />
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
					<p className="p-6 text-center text-xs text-muted-foreground">
						{query.trim() === ""
							? "No history yet — send a request and it will be recorded here."
							: "Nothing matches this search."}
					</p>
				) : (
					<table className="w-full border-separate border-spacing-0 text-left text-xs select-none">
						<thead>
							<tr className="font-mono text-[11px] uppercase tracking-wider text-muted-foreground">
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-3 py-2 font-medium">Time</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-2 py-2 font-medium">Method</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-3 py-2 font-medium">URL / Path</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-2 py-2 font-medium">Status</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-2 py-2 font-medium">Duration</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-2 py-2 font-medium">Env</th>
								<th className="sticky top-0 z-10 border-b border-border bg-card/80 px-3 py-2 text-right font-medium">Actions</th>
							</tr>
						</thead>
						<tbody>
							{displayEntries.map((entry) => (
								<tr key={entry.id} className="transition-colors hover:bg-muted/40">
									<td className="border-b border-border/50 px-3 py-1.5 font-mono text-[11px] tabular-nums text-muted-foreground">
										{formatTime(entry.createdAt)}
									</td>
									<td className="border-b border-border/50 px-2 py-1.5">
										<MethodLabel method={entry.method} />
									</td>
									<td className="max-w-[340px] border-b border-border/50 px-3 py-1.5">
										<span className="block truncate font-mono text-[11px] text-foreground/90" title={entry.url}>
											{entry.url || entry.requestPath || "—"}
										</span>
									</td>
									<td className="border-b border-border/50 px-2 py-1.5">
										<StatusPill status={entry.status} />
									</td>
									<td className="border-b border-border/50 px-2 py-1.5 font-mono text-[11px] tabular-nums text-muted-foreground">
										{formatDuration(entry.durationMs)}
									</td>
									<td className="border-b border-border/50 px-2 py-1.5 font-mono text-[11px] text-muted-foreground">
										{entry.env ? (
											<span className="rounded bg-muted/60 px-1 py-0.5 text-[10px]">{entry.env}</span>
										) : (
											"—"
										)}
									</td>
									<td className="border-b border-border/50 px-3 py-1.5 text-right">
										<div className="flex justify-end gap-1">
											<Button size="xs" variant="ghost" onClick={() => onReplay(entry)} title={`Replay ${entry.method} ${entry.url}`} className="gap-1 font-mono text-[10px]">
												<RotateCcw className="size-2.5" aria-hidden />
												Replay
											</Button>
											<Button size="xs" variant="ghost" onClick={() => { setReplayVarsId(entry.id); setReplayVars([{ key: "", value: "", enabled: true }]); }} title="Replay with vars" className="font-mono text-[10px]">
												Vars
											</Button>
										</div>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				)}
			</div>

			<AlertDialog open={!!replayVarsId} onOpenChange={(open) => { if (!open) setReplayVarsId(null); }}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Replay with vars</AlertDialogTitle>
						<AlertDialogDescription>Override request variables for this replay (e.g. token, id). Empty keys are dropped.</AlertDialogDescription>
					</AlertDialogHeader>
					<div className="max-h-64 overflow-auto">
						<KeyValueEditor rows={replayVars} onChange={setReplayVars} keyPlaceholder="var" valuePlaceholder="value" />
					</div>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction onClick={onReplayWithVars}>Replay</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
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
	);
}
