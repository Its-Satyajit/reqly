import { useEffect, useMemo, useState } from "react";
import { ScrollText, Trash2, Download, Filter, ShieldAlert } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
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
import { cn } from "#lib/utils";
import { copyText } from "#lib/response";
import { getAuditBridge, formatAuditTime } from "#lib/audit";
import type { AuditEntry } from "#lib/audit";

export function AuditView() {
	const [entries, setEntries] = useState<AuditEntry[]>([]);
	const [filter, setFilter] = useState("");
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [confirmClear, setConfirmClear] = useState(false);
	const [busy, setBusy] = useState(false);

	const load = async () => {
		setLoading(true);
		setError(null);
		try {
			const list = await getAuditBridge().list();
			setEntries(list ?? []);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock ledger for dev preview
				const now = Date.now();
				setEntries([
					{ id: "1", timestamp: new Date(now - 1000 * 60 * 12).toISOString(), actor: "local", action: "request.send", resource: "collections/api/users", details: "GET /users" },
					{ id: "2", timestamp: new Date(now - 1000 * 60 * 5).toISOString(), actor: "local", action: "auth.login", resource: "oauth2:client", details: "client_credentials" },
					{ id: "3", timestamp: new Date(now - 1000 * 30).toISOString(), actor: "local", action: "workflow.run", resource: "workflows/e2e.yaml", details: "3 steps" },
				]);
			} else setError(msg);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void load();
	}, []);

	const filtered = useMemo(() => {
		if (!filter.trim()) return entries;
		const q = filter.trim().toLowerCase();
		return entries.filter((e) => `${e.action} ${e.resource} ${e.actor} ${e.details ?? ""}`.toLowerCase().includes(q));
	}, [entries, filter]);

	const clear = async () => {
		setBusy(true);
		try {
			await getAuditBridge().clear();
			setEntries([]);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) setEntries([]);
			else setError(msg);
		} finally {
			setBusy(false);
			setConfirmClear(false);
		}
	};

	const doExport = async () => {
		try {
			const txt = await getAuditBridge().export();
			if (!txt) {
				// Mock export
				const mock = filtered.map((e) => `${e.timestamp} ${e.action} ${e.resource} ${e.details ?? ""}`).join("\n");
				await copyText(mock || "no entries");
				return;
			}
			await copyText(txt);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				const mock = filtered.map((e) => `${e.timestamp} ${e.action} ${e.resource}`).join("\n");
				await copyText(mock);
			} else setError(msg);
		}
	};

	return (
		<section className="flex h-full min-h-0 flex-col" aria-label="Audit log">
			<PageHeader
				icon={ScrollText}
				title="Audit Log"
				description="Append-only ledger — local, 0600, per-workspace .reqly/audit.log. No cloud, no reorder."
				actions={
					<div className="flex items-center gap-1.5">
						<Badge variant="outline" className="font-mono text-[10px]">{entries.length} entries</Badge>
						<Button variant="outline" size="xs" onClick={() => void doExport()} className="gap-1.5">
							<Download className="size-3.5" /> Export
						</Button>
						<Button variant="ghost" size="xs" onClick={() => setConfirmClear(true)} disabled={busy || entries.length === 0} className="gap-1.5 text-status-error hover:text-status-error">
							<Trash2 className="size-3.5" /> Clear
						</Button>
					</div>
				}
			/>

			<div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/20 px-3 py-2">
				<Filter className="size-3.5 text-muted-foreground" aria-hidden />
				<Input value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="filter action, resource, actor" className="h-7 w-64 font-mono text-xs" />
				<span className="font-mono text-[11px] text-muted-foreground">{filtered.length} / {entries.length}</span>
				<span className="ml-auto hidden items-center gap-1 font-mono text-[10px] text-muted-foreground sm:inline-flex"><ShieldAlert className="size-3" /> append-only · 0600</span>
			</div>

			{error && (
				<Alert variant="destructive" className="m-3 py-2">
					<AlertDescription className="font-mono text-xs">{error}</AlertDescription>
				</Alert>
			)}

			<div
				className="min-h-0 flex-1 overflow-auto p-2"
				aria-label="Audit ledger"
				style={{
					backgroundImage: `repeating-linear-gradient(to bottom, color-mix(in srgb, var(--border) 7%, transparent) 0 1px, transparent 1px 28px)`,
				}}
			>
				{loading ? (
					<p className="p-6 font-mono text-xs text-muted-foreground">Loading ledger…</p>
				) : filtered.length === 0 ? (
					<div className="flex min-h-[200px] flex-col items-center justify-center gap-2 rounded border border-dashed border-border bg-card/60 px-4 py-10 text-center">
						<ScrollText className="size-5 text-muted-foreground/50" aria-hidden />
						<p className="text-sm font-medium">No audit entries yet.</p>
						<p className="max-w-[52ch] text-balance font-mono text-xs text-muted-foreground">Every sensitive action appends a line — request.send, workflow.run, auth.login — with actor, resource, and ISO timestamp. Clear requires confirmation.</p>
					</div>
				) : (
					<div className="overflow-auto rounded border border-border bg-card">
						<table className="w-full border-collapse text-xs">
							<thead className="sticky top-0 z-10 border-b border-border bg-muted/30 backdrop-blur">
								<tr className="text-left font-mono text-[11px] text-muted-foreground">
									<th className="px-2.5 py-1.5 font-semibold">Time</th>
									<th className="px-2.5 py-1.5 font-semibold">Actor</th>
									<th className="px-2.5 py-1.5 font-semibold">Action</th>
									<th className="px-2.5 py-1.5 font-semibold">Resource</th>
									<th className="px-2.5 py-1.5 font-semibold">Details</th>
								</tr>
							</thead>
							<tbody>
								{filtered.map((e) => (
									<tr key={e.id} className="border-b border-border/40 last:border-0 hover:bg-muted/20">
										<td className="whitespace-nowrap px-2.5 py-1.5 font-mono text-[11px] tabular-nums text-muted-foreground">{formatAuditTime(e.timestamp)}</td>
										<td className="px-2.5 py-1.5">
											<Badge variant="outline" className="font-mono text-[10px]">{e.actor}</Badge>
										</td>
										<td className="px-2.5 py-1.5">
											<span className={cn("inline-flex rounded border px-1.5 py-0.5 font-mono text-[11px] font-medium", e.action.includes("auth") ? "border-status-redirect/20 bg-status-redirect/10 text-status-redirect" : e.action.includes("request") ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : "border-border bg-muted/40")}>{e.action}</span>
										</td>
										<td className="max-w-[28ch] truncate px-2.5 py-1.5 font-mono text-[11px]">{e.resource}</td>
										<td className="max-w-[32ch] truncate px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground" title={e.details ?? ""}>{e.details ?? "—"}</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</div>

			<div className="flex items-center gap-2 border-t border-border bg-muted/10 px-3 py-1.5 font-mono text-[10px] text-muted-foreground">
				<span>Ledger .reqly/audit.log · {entries.length} total</span>
				<span className="ml-auto">JSONL, 0600, local only</span>
			</div>

			<AlertDialog open={confirmClear} onOpenChange={setConfirmClear}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Clear audit log?</AlertDialogTitle>
						<AlertDialogDescription className="font-mono text-xs">This truncates .reqly/audit.log for this workspace. The file is append-only by convention — clearing is destructive and cannot be undone.</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Keep</AlertDialogCancel>
						<AlertDialogAction onClick={() => void clear()} className="bg-status-error text-white hover:bg-status-error/90">Clear</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</section>
	);
}
