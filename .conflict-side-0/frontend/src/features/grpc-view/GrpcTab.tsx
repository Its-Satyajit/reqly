import { useEffect, useRef } from "react";
import { Cable, Search, Send, Square } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { useGrpcStore } from "#stores/useGrpcStore";

function statusBadge(status: string) {
	switch (status) {
		case "unary-ok":
		case "done":
			return <Badge variant="secondary" className="text-status-ok">{status}</Badge>;
		case "streaming":
			return (
				<Badge variant="secondary" className="flex items-center gap-1">
					<Spinner className="size-2.5" /> streaming
				</Badge>
			);
		case "connecting":
			return (
				<Badge variant="secondary" className="flex items-center gap-1">
					<Spinner className="size-2.5" /> sending
				</Badge>
			);
		case "error":
			return <Badge variant="destructive">error</Badge>;
		default:
			return <Badge variant="ghost">idle</Badge>;
	}
}

/** GrpcTab is the desktop gRPC surface (M43 T7): reflection-driven
 * service/method picking, JSON message editing, unary responses and live
 * streaming inspectors. */
export function GrpcTab({ tabId }: { tabId: string }) {
	const tab = useGrpcStore((s) => s.tabs[tabId]);
	const update = useGrpcStore((s) => s.update);
	const discover = useGrpcStore((s) => s.discover);
	const send = useGrpcStore((s) => s.send);
	const stop = useGrpcStore((s) => s.stop);
	const newTab = useGrpcStore((s) => s.newTab);
	const logRef = useRef<HTMLDivElement>(null);

	// Lazy-init the persistent gRPC tab state on first render.
	useEffect(() => {
		newTab("grpc");
	}, [newTab]);

	useEffect(() => {
		if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
	}, [tab?.streamMessages.length]);

	if (!tab) return null;

	const selected = tab.services.find((svc) => svc.name === tab.service);
	const methods = selected?.methods ?? [];
	const isStreaming = methods.find((m) => m.name === tab.method)?.serverStreaming ?? false;
	const busy = tab.status === "connecting" || tab.status === "streaming";

	return (
		<div className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="gRPC Client">
			<PageHeader
				icon={Cable}
				title="gRPC Client"
				description="Server reflection, unary RPCs, and client/server streaming inspections"
			/>
			<div className="flex flex-col gap-2 p-3">
			<div className="flex items-center gap-1.5">
				<Input
					value={tab.target}
					onChange={(e) => update(tabId, { target: e.target.value })}
					placeholder="localhost:50051"
					spellCheck={false}
					className="flex-1 font-mono text-xs"
					aria-label="gRPC endpoint (host:port)"
				/>
				<Button variant="outline" size="sm" onClick={() => void discover(tabId)} disabled={tab.target.trim() === "" || busy}>
					{tab.status === "connecting" ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
					Discover
				</Button>
				{statusBadge(tab.status)}
			</div>

			<div className="flex items-center gap-1.5">
				<select
					value={tab.service}
					onChange={(e) => update(tabId, { service: e.target.value, method: "" })}
					disabled={tab.services.length === 0 || busy}
					aria-label="Service"
					className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
				>
					<option value="">Service…</option>
					{tab.services.map((s) => (
						<option key={s.name} value={s.name}>{s.name}</option>
					))}
				</select>
				<select
					value={tab.method}
					onChange={(e) => update(tabId, { method: e.target.value })}
					disabled={!tab.service || busy}
					aria-label="Method"
					className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
				>
					<option value="">Method…</option>
					{methods.map((m) => (
						<option key={m.fullName} value={m.name}>
							{m.name}{m.serverStreaming ? " (stream)" : ""}
						</option>
					))}
				</select>
				<label className="ml-auto flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
					<input
						type="checkbox"
						checked={tab.tlsSkipVerify}
						onChange={(e) => update(tabId, { tlsSkipVerify: e.target.checked, tls: e.target.checked || tab.tls })}
						aria-label="Use TLS (skip verification)"
						className="size-3.5 accent-(--primary)"
					/>
					TLS
				</label>
				<Input
					value={tab.protoFiles}
					onChange={(e) => update(tabId, { protoFiles: e.target.value })}
					placeholder="proto files (comma-separated, optional)"
					spellCheck={false}
					aria-label="Proto files fallback"
					className="w-64 font-mono text-xs"
				/>
			</div>

			{tab.error && (
				<Alert variant="destructive">
					<AlertDescription>{tab.error}</AlertDescription>
				</Alert>
			)}

			<div className="min-h-0 flex-[2] rounded-md border border-border">
				<div className="border-b border-border px-2 py-1 text-xs text-muted-foreground">Request message (JSON)</div>
				<CodeMirrorEditor
					value={tab.message}
					language="json"
					onChange={(v) => update(tabId, { message: v })}
					className="h-[calc(100%-1.75rem)] overflow-hidden"
				/>
			</div>

			<div className="flex items-center gap-1.5">
				<Button
					size="sm"
					onClick={() => void send(tabId)}
					disabled={!tab.service || !tab.method || busy || tab.message.trim() === ""}
				>
					<Send data-icon="inline-start" />
					Send
				</Button>
				{isStreaming && tab.status === "streaming" && (
					<Button size="sm" variant="destructive" onClick={() => void stop(tabId)}>
						<Square data-icon="inline-start" />
						Stop
					</Button>
				)}
				{!isStreaming && <span className="text-xs text-muted-foreground">Unary call through the shared pipeline.</span>}
				{isStreaming && <span className="text-xs text-muted-foreground">Server-streaming — messages stream below.</span>}
			</div>

			<div
				ref={logRef}
				className="min-h-0 flex-1 overflow-y-auto rounded-md border border-border p-2"
				aria-label="Response"
			>
				{tab.unaryResult?.messageJson ? (
					<CodeMirrorEditor value={tab.unaryResult.messageJson} language="json" readOnly className="h-full overflow-hidden" />
				) : tab.streamMessages.length === 0 ? (
					<p className="text-xs text-muted-foreground">No response yet.</p>
				) : (
					<ul className="flex flex-col gap-1">
						{tab.streamMessages.map((m) => (
							<li key={m.seq} className="rounded border border-border/60 px-2 py-1 font-mono text-[11px]">
								<span className="mr-2 text-muted-foreground">#{m.seq}</span>
								<span className="break-all whitespace-pre-wrap">{m.messageJson}</span>
							</li>
						))}
					</ul>
				)}
			</div>
			</div>
		</div>
	);
}

