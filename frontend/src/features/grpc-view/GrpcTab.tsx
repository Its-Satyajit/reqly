import { useEffect, useRef } from "react";
import { RefreshCw, Square, Zap } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { KeyValueEditor } from "#components/KeyValueEditor";
import { cn } from "#lib/utils";
import { useGrpcStore } from "#stores/useGrpcStore";
import { ViewShell } from "../../components/shell/ViewLayout";

/** statusBadge renders the connection state pill. */
function statusBadge(status: string) {
	if (status === "streaming" || status === "connecting") {
		return (
			<span className="flex shrink-0 items-center gap-1 rounded-full border border-border bg-muted/40 px-2 py-0.5 font-data text-2xs">
				<Spinner className="size-2.5" /> {status}
			</span>
		);
	}
	if (status === "error") {
		return (
			<span className="shrink-0 rounded-full border border-status-error/40 px-2 py-0.5 font-data text-2xs text-status-error">
				error
			</span>
		);
	}
	if (status === "unary-ok" || status === "done") {
		return (
			<span className="shrink-0 rounded-full border border-status-ok/40 px-2 py-0.5 font-data text-2xs text-status-ok">
				{status}
			</span>
		);
	}
	return null;
}

function FieldLabel({ children }: { children: string }) {
	return (
		<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
			{children}
		</p>
	);
}

function selectClass() {
	return "h-8 w-full rounded-md border border-border bg-transparent px-2 text-xs";
}

/** GrpcTab is the desktop gRPC surface (M43 T7, restyled G-17.4.3):
 * reflection-driven service/method picking, TLS/plaintext segmented toggle,
 * streaming-type + type chips, JSON message editor, metadata table with
 * deadline, orange Invoke, and a CHANNEL footer. */
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
	const method = selected?.methods.find((m) => m.name === tab.method);
	const methods = selected?.methods ?? [];
	const isStreaming = method?.serverStreaming ?? false;
	const busy = tab.status === "connecting" || tab.status === "streaming";
	const tlsOn = tab.tls || tab.tlsSkipVerify;

	return (
		<ViewShell label="gRPC tab" className="overflow-y-auto">
			<div className="flex items-center gap-2">
				<Input
					value={tab.target}
					onChange={(e) => update(tabId, { target: e.target.value })}
					placeholder="api.example.dev:443"
					spellCheck={false}
					className="flex-1 font-mono text-xs"
					aria-label="gRPC endpoint (host:port)"
				/>
				<div
					role="radiogroup"
					aria-label="Channel security"
					className="flex shrink-0 items-center rounded-full border border-border p-0.5"
				>
					{(
						[
							{ label: "TLS", on: tlsOn, next: () => update(tabId, { tls: true }) },
							{
								label: "plaintext",
								on: !tlsOn,
								next: () => update(tabId, { tls: false, tlsSkipVerify: false }),
							},
						] as const
					).map((opt) => (
						<button
							key={opt.label}
							type="button"
							role="radio"
							aria-checked={opt.on}
							onClick={opt.next}
							className={cn(
								"rounded-full px-2.5 py-0.5 font-data text-2xs",
								opt.on
									? "bg-primary/15 text-primary"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							{opt.label}
						</button>
					))}
				</div>
				<Button
					size="sm"
					variant="outline"
					className="shrink-0"
					onClick={() => void discover(tabId)}
					disabled={tab.target.trim() === "" || busy}
				>
					{tab.status === "connecting" ? (
						<Spinner data-icon="inline-start" />
					) : (
						<RefreshCw data-icon="inline-start" />
					)}
					Reflect
				</Button>
				{statusBadge(tab.status)}
			</div>

			<div className="flex flex-col gap-1">
				<FieldLabel>Service</FieldLabel>
				<select
					value={tab.service}
					onChange={(e) => update(tabId, { service: e.target.value, method: "" })}
					disabled={tab.services.length === 0 || busy}
					aria-label="Service"
					className={selectClass()}
				>
					<option value="">Service…</option>
					{tab.services.map((s) => (
						<option key={s.name} value={s.name}>
							{s.name}
						</option>
					))}
				</select>
			</div>

			<div className="flex flex-col gap-1">
				<FieldLabel>Method</FieldLabel>
				<select
					value={tab.method}
					onChange={(e) => update(tabId, { method: e.target.value })}
					disabled={!tab.service || busy}
					aria-label="Method"
					className={selectClass()}
				>
					<option value="">Method…</option>
					{methods.map((m) => (
						<option key={m.fullName} value={m.name}>
							{m.name}
						</option>
					))}
				</select>
			</div>

			{method ? (
				<div className="flex flex-wrap items-center gap-1.5">
					<span className="rounded-full border border-status-ok/40 px-2 py-0.5 font-data text-2xs text-status-ok">
						{isStreaming ? "server stream" : "unary stream"}
					</span>
					<span className="rounded-full border border-border bg-muted/30 px-2 py-0.5 font-data text-2xs text-muted-foreground">
						{method.inputType} → {method.outputType}
					</span>
				</div>
			) : null}

			<div className="grid shrink-0 grid-cols-1 gap-3 lg:grid-cols-[3fr_2fr]">
				<div className="flex min-h-0 flex-col gap-1">
					<FieldLabel>
						{`Message${method ? ` (${method.inputType})` : ""}`}
					</FieldLabel>
					<div className="h-44 overflow-hidden rounded-lg border border-border">
						<CodeMirrorEditor
							value={tab.message}
							language="json"
							onChange={(v) => update(tabId, { message: v })}
							className="h-full"
						/>
					</div>
				</div>
				<div className="flex min-h-0 flex-col gap-1">
					<FieldLabel>Metadata</FieldLabel>
					<div className="max-h-44 overflow-y-auto rounded-lg border border-border bg-background p-2">
						<KeyValueEditor
							rows={tab.metadata}
							onChange={(rows) => update(tabId, { metadata: rows })}
							keyPlaceholder="Key"
							valuePlaceholder="Value"
						/>
						{tab.metadata.length === 0 ? (
							<p className="pt-1 text-center text-xs text-muted-foreground italic">
								This table is empty.
							</p>
						) : null}
					</div>
					<div className="flex items-center gap-2 text-xs text-muted-foreground">
						<span>deadline:</span>
						<Input
							value={tab.deadline}
							onChange={(e) => update(tabId, { deadline: e.target.value })}
							placeholder="15s"
							aria-label="Deadline"
							spellCheck={false}
							className="h-6 w-20 px-1.5 font-mono text-xs"
						/>
						<span>· compression: identity</span>
					</div>
				</div>
			</div>

			{tab.error ? (
				<Alert variant="destructive">
					<AlertDescription>{tab.error}</AlertDescription>
				</Alert>
			) : null}

			<div className="flex items-center gap-2">
				<Button
					size="sm"
					variant="destructive"
					onClick={() => void send(tabId)}
					disabled={!tab.service || !tab.method || busy || tab.message.trim() === ""}
				>
					<Zap data-icon="inline-start" />
					Invoke
				</Button>
				{isStreaming && tab.status === "streaming" ? (
					<Button size="sm" variant="outline" onClick={() => void stop(tabId)}>
						<Square data-icon="inline-start" />
						Stop
					</Button>
				) : null}
				<Input
					value={tab.protoFiles}
					onChange={(e) => update(tabId, { protoFiles: e.target.value })}
					placeholder="proto files (comma-separated, optional fallback)"
					spellCheck={false}
					aria-label="Proto files fallback"
					className="ml-auto w-64 font-mono text-xs"
				/>
			</div>

			<div className="flex flex-col gap-1">
				<FieldLabel>Responses</FieldLabel>
				<div
					ref={logRef}
					className="flex min-h-40 flex-1 flex-col overflow-y-auto rounded-lg border border-border bg-card"
				>
					{tab.unaryResult?.messageJson ? (
						<CodeMirrorEditor
							value={tab.unaryResult.messageJson}
							language="json"
							readOnly
							className="h-full overflow-hidden"
						/>
					) : tab.streamMessages.length > 0 ? (
						<ul className="flex flex-col gap-1 p-2">
							{tab.streamMessages.map((m) => (
								<li
									key={m.seq}
									className="rounded border border-border/60 px-2 py-1 font-mono text-xs"
								>
									<span className="mr-2 text-muted-foreground">#{m.seq}</span>
									<span className="break-all whitespace-pre-wrap">{m.messageJson}</span>
								</li>
							))}
						</ul>
					) : (
						<div className="flex flex-1 flex-col items-center justify-center gap-1.5 py-8">
							<Zap className="size-8 text-muted-foreground/40" aria-hidden />
							<p className="text-sm font-medium text-foreground">No messages yet</p>
							<p className="text-xs text-muted-foreground">
								Invoke a unary call or open a stream.
							</p>
						</div>
					)}
				</div>
			</div>

			<div className="flex shrink-0 flex-col gap-1 border-t border-border pt-3">
				<FieldLabel>Channel</FieldLabel>
				<dl className="grid grid-cols-[7rem_1fr] gap-x-3 gap-y-0.5 font-data text-xs">
					<dt className="text-muted-foreground">ADDRESS</dt>
					<dd className="text-foreground">{tab.target.trim() || "—"}</dd>
					<dt className="text-muted-foreground">SECURITY</dt>
					<dd className="text-foreground">{tlsOn ? "TLS" : "plaintext"}</dd>
					<dt className="text-muted-foreground">USER-AGENT</dt>
					<dd className="text-foreground">reqly-grpc/1.4</dd>
				</dl>
			</div>
		</ViewShell>
	);
}
