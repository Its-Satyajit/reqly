import { useEffect, useRef, useState } from "react";
import { Activity, CircleStop, PlugZap, Trash2 } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import type { KeyValueRow } from "#lib/request";
import { cn } from "#lib/utils";
import { bytesToBase64 } from "#lib/response";
import { KeyValueEditor } from "#components/KeyValueEditor";
import {
	formatFrameTime,
	type RealtimeFrameView,
	type RealtimeKind,
} from "#lib/realtime";
import { useRealtimeStore } from "#stores/useRealtimeStore";

/** statusPill renders the connection-state pill (CLOSED/IDLE/…). */
function statusPill(status: string) {
	const tone =
		status === "connected"
			? "border-status-ok/40 text-status-ok"
			: status === "error" || status === "closed"
				? "border-status-error/50 text-status-error"
				: "border-border text-muted-foreground";
	const label = status === "idle" ? "idle" : status.toUpperCase();
	return (
		<span
			className={cn(
				"flex shrink-0 items-center gap-1 rounded-full border px-2 py-0.5 font-data text-2xs font-semibold",
				tone,
			)}
		>
			{status === "connecting" ? <Spinner className="size-2.5" /> : null}
			{label}
		</span>
	);
}

function CardTitle({ children, right }: { children: React.ReactNode; right?: React.ReactNode }) {
	return (
		<div className="flex shrink-0 items-center justify-between gap-2">
			<h3 className="text-xs font-semibold text-foreground">{children}</h3>
			{right}
		</div>
	);
}

/** FrameRow is one message-log line: time, direction, SSE name/id/retry, data. */
function FrameRow({ frame, isWS }: { frame: RealtimeFrameView; isWS: boolean }) {
	return (
		<li
			className={cn(
				"rounded border border-border/50 px-2 py-1 font-mono text-xs",
				frame.direction === "out" && "bg-muted/40",
			)}
		>
			<span className="mr-2 text-muted-foreground">{formatFrameTime(frame.timestamp)}</span>
			<span
				className={cn(
					"mr-2 font-sans font-semibold",
					frame.direction === "out" ? "text-status-info" : "text-status-ok",
				)}
			>
				{frame.direction === "out" ? "↑ out" : isWS ? "↓ in" : "event"}
			</span>
			{frame.name || frame.id ? (
				<span className="mr-2 font-sans text-muted-foreground">
					{frame.name || "message"}
					{frame.id ? ` #${frame.id}` : ""}
				</span>
			) : null}
			{frame.retryMs != null && frame.retryMs > 0 ? (
				<span
					className="mr-2 rounded bg-muted px-1 font-sans text-muted-foreground"
					title='Server "retry:" field — suggested reconnect delay'
				>
					retry {frame.retryMs}ms
				</span>
			) : null}
			<span className="break-all whitespace-pre-wrap">
				{frame.encoding === "base64" ? `(binary ${frame.data?.length ?? 0} b64) ` : ""}
				{frame.data}
			</span>
		</li>
	);
}

/** MessageLog renders frames with an empty state; pause freezes rendering. */
function MessageLog({
	frames,
	isWS,
	paused,
	pausedFrames,
	autoScroll,
	logRef,
	emptyTitle,
	emptyHint,
}: {
	frames: RealtimeFrameView[];
	isWS: boolean;
	paused: boolean;
	pausedFrames: RealtimeFrameView[];
	autoScroll: boolean;
	logRef: React.RefObject<HTMLUListElement | null>;
	emptyTitle: string;
	emptyHint: string;
}) {
	const shown = paused ? pausedFrames : frames;
	if (shown.length === 0) {
		return (
			<div className="flex flex-1 flex-col items-center justify-center gap-1.5 py-10">
				{isWS ? (
					<PlugZap className="size-8 text-muted-foreground/40" aria-hidden />
				) : (
					<Activity className="size-8 text-muted-foreground/40" aria-hidden />
				)}
				<p className="text-sm font-medium text-foreground">{emptyTitle}</p>
				<p className="max-w-xs text-center text-xs text-muted-foreground">{emptyHint}</p>
			</div>
		);
	}
	return (
		<ul
			ref={logRef}
			className={cn("flex flex-1 flex-col gap-1 overflow-y-auto p-2", !autoScroll && "overflow-hidden")}
		>
			{shown.map((f) => (
				<FrameRow key={f.seq ?? f.timestamp} frame={f} isWS={isWS} />
			))}
		</ul>
	);
}

/** RealtimeTab is the G-17.4.4 realtime surface (WebSocket + SSE), restyled
 * to the design-9 frames: WS gets a two-pane composer/log layout with a
 * connection-info card; SSE gets an event-type filter, pause, auto-scroll,
 * and retry/last-event-id readout. */
export function RealtimeTab({ tabId }: { tabId: string }) {
	const tab = useRealtimeStore((s) => s.tabs[tabId]);
	const update = useRealtimeStore((s) => s.update);
	const connect = useRealtimeStore((s) => s.connect);
	const send = useRealtimeStore((s) => s.send);
	const sendBinary = useRealtimeStore((s) => s.sendBinary);
	const disconnect = useRealtimeStore((s) => s.disconnect);
	const clearFrames = useRealtimeStore((s) => s.clearFrames);
	const [draft, setDraft] = useState("");
	const [binaryMode, setBinaryMode] = useState(false);
	const [headersOpen, setHeadersOpen] = useState(false);
	const [paused, setPaused] = useState(false);
	const [pausedFrames, setPausedFrames] = useState<RealtimeFrameView[]>([]);
	const [autoScroll, setAutoScroll] = useState(true);
	const [eventFilter, setEventFilter] = useState("");
	const logRef = useRef<HTMLUListElement>(null);

	// Keep the log pinned to the newest frame while streaming and scrolling.
	useEffect(() => {
		if (autoScroll && logRef.current && !paused) {
			logRef.current.scrollTop = logRef.current.scrollHeight;
		}
	}, [tab?.frames.length, autoScroll, paused]);

	if (!tab) return null;
	const isWS = tab.kind === "ws";
	const connected = tab.status === "connected";

	const pause = (): void => {
		if (!paused) setPausedFrames(tab.frames);
		setPaused(!paused);
	};

	// SSE event-type filter: distinct frame names, newest last.
	const eventNames = isWS
		? []
		: [...new Set(tab.frames.map((f) => f.name).filter((n): n is string => Boolean(n)))];
	const visibleFrames = eventFilter
		? tab.frames.filter((f) => f.name === eventFilter)
		: tab.frames;
	const lastRetry = visibleFrames.filter((f) => f.retryMs).at(-1)?.retryMs;
	const lastEventId = visibleFrames.filter((f) => f.id).at(-1)?.id;

	return (
		<div className="flex h-full min-h-0 flex-col gap-3 p-4">
			<div className="flex items-center gap-2">
				{statusPill(tab.status)}
				<select
					value={tab.kind}
					disabled={connected || tab.status === "connecting"}
					onChange={(e) => {
						// SAFETY: options are exactly the two supported realtime kinds.
						update(tabId, { kind: e.target.value as RealtimeKind });
					}}
					aria-label="Protocol"
					className="h-8 shrink-0 rounded-md border border-border bg-transparent px-2 text-xs"
				>
					<option value="ws">WebSocket</option>
					<option value="sse">SSE</option>
				</select>
				<Input
					value={tab.url}
					onChange={(e) => update(tabId, { url: e.target.value })}
					placeholder={isWS ? "wss://stream.example.dev/v1/events" : "https://example.com/v1/events/stream"}
					spellCheck={false}
					className="flex-1 font-mono text-xs"
					aria-label="Endpoint URL"
				/>
				{connected || tab.status === "connecting" ? (
					<Button variant="outline" size="sm" onClick={() => void disconnect(tabId)}>
						{tab.status === "connecting" ? (
							<Spinner data-icon="inline-start" />
						) : (
							<CircleStop data-icon="inline-start" />
						)}
						Disconnect
					</Button>
				) : (
					<Button variant="destructive" size="sm" onClick={() => void connect(tabId)} disabled={tab.url.trim() === ""}>
						<PlugZap data-icon="inline-start" />
						Connect
					</Button>
				)}
			</div>

			<button
				type="button"
				className="flex w-fit items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
				aria-expanded={headersOpen}
				onClick={() => setHeadersOpen(!headersOpen)}
			>
				Headers {tab.headers.length > 0 && `(${tab.headers.filter((h) => h.enabled).length})`}
			</button>
			{headersOpen ? (
				<div
					className={cn("pt-1", (connected || tab.status === "connecting") && "pointer-events-none opacity-50")}
					title={connected ? "Disconnect before editing headers" : undefined}
				>
					<KeyValueEditor
						rows={tab.headers}
						onChange={(rows: KeyValueRow[]) => update(tabId, { headers: rows })}
						keyPlaceholder="header name"
						valuePlaceholder="value"
					/>
				</div>
			) : null}

			{tab.error ? (
				<Alert variant="destructive">
					<AlertDescription>{tab.error}</AlertDescription>
				</Alert>
			) : null}

			{isWS ? (
				<div className="grid min-h-0 flex-1 grid-cols-1 gap-3 lg:grid-cols-2">
					<div className="flex min-h-0 flex-col gap-2 rounded-xl border border-border bg-card p-3">
						<CardTitle>Message composer</CardTitle>
						<div className="h-36 shrink-0 overflow-hidden rounded-lg border border-border">
							<CodeMirrorDraft
								draft={draft}
								setDraft={setDraft}
								binaryMode={binaryMode}
							/>
						</div>
						<div className="flex flex-wrap items-center gap-1.5">
							<Button
								size="sm"
								variant="destructive"
								onClick={() => {
									if (draft.trim() === "" || !connected) return;
									if (binaryMode) {
										void sendBinary(tabId, bytesToBase64(draft));
									} else {
										void send(tabId, draft);
									}
									setDraft("");
								}}
								disabled={draft.trim() === "" || !connected}
							>
								Send
							</Button>
							<Button
								size="sm"
								variant="outline"
								onClick={() => {
									if (!connected) return;
									void send(tabId, binaryMode ? draft : JSON.stringify({ type: "ping" }));
									setDraft("");
								}}
								disabled={!connected}
							>
								Ping
							</Button>
							<Button
								size="sm"
								variant="outline"
								onClick={() => {
									if (!connected) return;
									void send(tabId, '{"type": ');
								}}
								disabled={!connected}
								title="Send intentionally invalid JSON to test error handling"
							>
								Send malformed JSON
							</Button>
						</div>
						{!connected ? (
							<p className="flex items-center gap-1.5 rounded-lg border border-status-warn/40 bg-status-warn/10 px-2.5 py-1.5 text-xs text-status-warn">
								ⓘ Socket is closed — connect to enable sending.
							</p>
						) : null}
						<label className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
							<input
								type="checkbox"
								checked={binaryMode}
								onChange={(e) => setBinaryMode(e.target.checked)}
								aria-label="Send as binary frame (UTF-8 bytes)"
								className="size-3.5 accent-(--primary)"
							/>
							binary frame
						</label>
					</div>
					<div className="flex min-h-0 flex-col rounded-xl border border-border bg-card p-3">
						<CardTitle
							right={
								<Button size="xs" variant="ghost" onClick={() => clearFrames(tabId)}>
									<Trash2 data-icon="inline-start" />
									Clear
								</Button>
							}
						>
							Message log{" "}
							<span className="ml-1 rounded bg-muted px-1.5 font-data text-2xs text-muted-foreground">
								{(paused ? pausedFrames : tab.frames).length}
							</span>
						</CardTitle>
						<MessageLog
							frames={tab.frames}
							isWS
							paused={paused}
							pausedFrames={pausedFrames}
							autoScroll={autoScroll}
							logRef={logRef}
							emptyTitle="No traffic yet"
							emptyHint="Frames you send and events pushed by the server appear here."
						/>
					</div>
				</div>
			) : (
				<div className="flex min-h-0 flex-1 flex-col gap-2 rounded-xl border border-border bg-card p-3">
					<div className="flex flex-wrap items-center gap-2">
						<span className="text-xs text-muted-foreground">Filter event type</span>
						<select
							value={eventFilter}
							onChange={(e) => setEventFilter(e.target.value)}
							aria-label="Filter event type"
							className="h-7 rounded-md border border-border bg-transparent px-2 text-xs"
						>
							<option value="">{`all (${tab.frames.length})`}</option>
							{eventNames.map((n) => (
								<option key={n} value={n}>
									{n}
								</option>
							))}
						</select>
						<Button size="sm" variant="outline" onClick={pause} disabled={tab.frames.length === 0}>
							{paused ? "Resume stream" : "Pause stream"}
						</Button>
						<Button size="sm" variant="ghost" onClick={() => clearFrames(tabId)}>
							<Trash2 data-icon="inline-start" />
							Clear
						</Button>
						<label className="flex items-center gap-1 text-xs text-muted-foreground">
							<input
								type="checkbox"
								checked={autoScroll}
								onChange={(e) => setAutoScroll(e.target.checked)}
								className="size-3.5 accent-(--primary)"
							/>
							auto-scroll
						</label>
						<span className="ml-auto font-data text-xs text-muted-foreground">
							retry: {lastRetry != null ? `${lastRetry}ms` : "—"} · last-event-id:{" "}
							{lastEventId ?? "—"}
						</span>
					</div>
					<MessageLog
						frames={visibleFrames}
						isWS={false}
						paused={paused}
						pausedFrames={pausedFrames}
						autoScroll={autoScroll}
						logRef={logRef}
						emptyTitle="Stream is idle"
						emptyHint="Connect to an endpoint to watch server-sent events arrive in real time."
					/>
				</div>
			)}

			{isWS ? (
				<div className="flex shrink-0 flex-col gap-1 rounded-lg border border-border bg-card p-3">
					<p className="text-xs font-semibold text-foreground">Connection info</p>
					<dl className="grid grid-cols-[8rem_1fr] gap-x-3 gap-y-0.5 font-data text-xs">
						<dt className="text-muted-foreground">READYSTATE</dt>
						<dd className="text-foreground">{tab.status.toUpperCase()}</dd>
						<dt className="text-muted-foreground">HEADERS</dt>
						<dd className="text-foreground">
							{tab.headers.filter((h) => h.enabled).length} set
						</dd>
					</dl>
				</div>
			) : null}
		</div>
	);
}

/** CodeMirrorDraft isolates the composer editor (kept out of the main
 * component so the grid layout stays readable). */
function CodeMirrorDraft({
	draft,
	setDraft,
	binaryMode,
}: {
	draft: string;
	setDraft: (v: string) => void;
	binaryMode: boolean;
}) {
	return (
		<CodeMirrorEditor
			value={draft}
			onChange={setDraft}
			language={binaryMode ? "text" : "json"}
			className="h-full"
		/>
	);
}
