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
import { HISTORY_PAGE_SIZE, useHistoryStore } from "../../stores/useHistoryStore";
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
	}, [load]);

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

	return (
		<div className="flex h-full min-h-0 flex-col p-4">
			<div className="flex shrink-0 flex-col gap-3 pb-3">
				<div>
					<h2 className="text-sm font-semibold">History</h2>
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
						className="min-w-0 flex-1 rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground"
					/>
					<Button size="sm" variant="outline" onClick={search} disabled={loading}>
						<Search className="size-3.5" aria-hidden />
						<span className="sr-only">Search</span>
					</Button>
				</div>
				<div className="flex items-center gap-2">
					<CompactSelect
						value={status}
						onChange={(next) => {
							setStatus(next);
							setPage(0);
							void load({ offset: 0, status: next, query });
						}}
						ariaLabel="Status filter"
						options={[
							{ value: "", label: "All statuses" },
							{ value: "2xx", label: "2xx" },
							{ value: "4xx", label: "4xx" },
							{ value: "5xx", label: "5xx" },
						]}
					/>
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
				{entries.length === 0 && !loading ? (
					<p className="p-6 text-center text-xs text-muted-foreground">
						No history yet — send a request and it will be recorded here.
					</p>
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
							{entries.map((entry) => (
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
											<RotateCcw className="size-3" aria-hidden />
											Replay
										</Button>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				)}
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
	);
}
