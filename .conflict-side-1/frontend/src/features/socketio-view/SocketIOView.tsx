import { useEffect, useMemo, useRef, useState } from "react";
import { Cable, Copy, Filter, Pause, Play, Radio, Trash2, Zap } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { cn } from "#lib/utils";
import { copyText } from "#lib/response";
import { isString } from "#lib/typeGuards";
import { getSocketIOBridge } from "#lib/socketio";
import type { SocketIOFrameView } from "#lib/socketio";

function formatTime(ts: number): string {
	const d = new Date(ts);
	return d.toLocaleTimeString(undefined, { hour12: false }) + "." + String(d.getMilliseconds()).padStart(3, "0");
}

function isJson(s: string): boolean {
	try {
		JSON.parse(s);
		return true;
	} catch {
		return false;
	}
}

export function SocketIOView() {
	const [url, setUrl] = useState("ws://localhost:3000");
	const [namespace, setNamespace] = useState("/");
	const [eventName, setEventName] = useState("chat:message");
	const [eventData, setEventData] = useState('{"text":"hello from Reqly"}');
	const [connected, setConnected] = useState(false);
	const [sessionId] = useState(() => `sio-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [ok, setOk] = useState<string | null>(null);

	const [frames, setFrames] = useState<SocketIOFrameView[]>([]);
	const [filter, setFilter] = useState("");
	const [paused, setPaused] = useState(false);
	const logRef = useRef<HTMLDivElement>(null);
	const offRef = useRef<(() => void) | null>(null);
	const mockTimerRef = useRef<number | null>(null);

	useEffect(() => {
		if (!connected) {
			offRef.current?.();
			offRef.current = null;
			if (mockTimerRef.current) {
				window.clearInterval(mockTimerRef.current);
				mockTimerRef.current = null;
			}
			return;
		}
		let cancelled = false;
		const bridge = getSocketIOBridge();
		const off = bridge.onFrame(sessionId, (f) => {
			if (paused || cancelled) return;
			if (f.type === "message") setFrames((prev) => [...prev.slice(-499), f]);
			if (f.type === "error") setError(f.raw ?? "socket error");
		});
		offRef.current = off;

		const start = (): void => {
			bridge
				.connect({ sessionId, url: url.trim(), namespace: namespace.trim() })
				.then(() => {
					if (cancelled) return;
					setError(null);
					setOk(`connected to ${namespace.trim() || "/"} @ ${url.trim()}`);
					setFrames((prev) => [
						...prev.slice(-499),
						{ sessionId, type: "status" as const, raw: `connected`, timestamp: Date.now() },
					]);
				})
				.catch((err) => {
					if (cancelled) return;
					const msg = err instanceof Error ? err.message : String(err);
					if (msg.includes("not available in this build")) {
						let n = 0;
						mockTimerRef.current = window.setInterval(() => {
							if (paused || cancelled) return;
							const evs = [
								{ event: "chat:message", data: { user: "bot", text: "welcome", n } },
								{ event: "presence:join", data: { user: `u${n % 5}`, at: Date.now() } },
								{ event: "chat:typing", data: { user: "alice" } },
							];
							const pick = evs[n % evs.length];
							setFrames((prev) => [
								...prev.slice(-499),
								{ sessionId, type: "message", namespace: namespace.trim() || "/", event: pick.event, data: pick.data, timestamp: Date.now() },
							]);
							n++;
						}, 1300);
						setError(null);
						setOk("connected (mock — no server in dev)");
						setFrames((prev) => [...prev.slice(-499), { sessionId, type: "status" as const, raw: `connected to ${url.trim()}${namespace.trim() || "/"}`, timestamp: Date.now() }]);
						return;
					}
					setError(msg);
					setConnected(false);
				});
		};
		start();

		return () => {
			cancelled = true;
			off();
			offRef.current = null;
			if (mockTimerRef.current) {
				window.clearInterval(mockTimerRef.current);
				mockTimerRef.current = null;
			}
			void getSocketIOBridge().close(sessionId).catch(() => {});
		};
	}, [connected, sessionId, url, namespace, paused]);

	useEffect(() => {
		if (logRef.current && !paused) logRef.current.scrollTop = logRef.current.scrollHeight;
	}, [frames.length, paused]);

	const filtered = useMemo(() => {
		if (!filter.trim()) return frames;
		const q = filter.trim().toLowerCase();
		return frames.filter((f) => {
			const blob = `${f.event ?? ""} ${JSON.stringify(f.data ?? "")} ${f.namespace ?? ""} ${f.raw ?? ""}`.toLowerCase();
			return blob.includes(q);
		});
	}, [frames, filter]);

	const emit = async () => {
		if (!url.trim() || !eventName.trim()) {
			setError("URL and event are required");
			return;
		}
		setBusy(true);
		setError(null);
		setOk(null);
		let data: unknown = eventData;
		if (eventData.trim() !== "") {
			try {
				data = JSON.parse(eventData);
			} catch {
				data = eventData;
			}
		}
		try {
			await getSocketIOBridge().emit({ sessionId, url: url.trim(), event: eventName.trim(), data, namespace: namespace.trim() });
			setOk(`emit → ${eventName.trim()}`);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setOk(`emit queued to ${eventName.trim()} (mock)`);
			} else {
				setError(msg);
				return;
			}
		} finally {
			setBusy(false);
		}
		if (!paused) {
			setFrames((prev) => [
				...prev.slice(-499),
				{ sessionId, type: "message", namespace: namespace.trim() || "/", event: eventName.trim(), data, timestamp: Date.now() },
			]);
		}
	};

	return (
		<section className="flex h-full min-h-0 flex-col overflow-hidden" aria-label="Socket.IO">
			<PageHeader
				icon={Cable}
				title="Socket.IO"
				description="Connect, listen, and emit over Socket.IO — namespaces, events, and the ledger."
				actions={
					<div className="flex items-center gap-1.5">
						<span className={cn("inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-mono", connected ? "border-status-ok/30 bg-status-ok/10 text-status-ok" : "border-border bg-muted/40 text-muted-foreground")}>
							<span className={cn("size-1.5 rounded-full", connected ? "bg-status-ok animate-pulse" : "bg-muted-foreground/40")} aria-hidden />
							{connected ? "connected" : "idle"}
						</span>
						<Badge variant="outline" className="font-mono text-[10px]">{frames.filter((f) => f.type === "message").length} events</Badge>
					</div>
				}
			/>

			<div className="flex min-h-0 flex-1 flex-col gap-0 overflow-hidden">
				<div className="shrink-0 border-b border-border bg-card/20">
					<div className="grid gap-0 lg:grid-cols-2">
						<div className="flex flex-col gap-2.5 border-b border-border p-3 lg:border-b-0 lg:border-r">
							<div className="flex items-center gap-2">
								<Radio className="size-3.5 text-muted-foreground" aria-hidden />
								<h3 className="text-xs font-semibold tracking-tight">Connect</h3>
								<span className="text-[11px] text-muted-foreground">— namespace + transport</span>
							</div>

							<label className="flex flex-col gap-1">
								<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">URL</span>
								<Input value={url} onChange={(e) => setUrl(e.target.value)} placeholder="ws://localhost:3000  or  https://chat.example.com" spellCheck={false} className="font-mono text-xs" disabled={connected} />
							</label>

							<div className="grid grid-cols-[1fr_auto] gap-2">
								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Namespace</span>
									<Input value={namespace} onChange={(e) => setNamespace(e.target.value)} placeholder="/  ·  /chat  ·  /presence" spellCheck={false} className="font-mono text-xs" disabled={connected} />
								</label>
								<div className="flex items-end">
									{connected ? (
										<Button variant="outline" size="sm" className="w-full" onClick={() => setConnected(false)}>Disconnect</Button>
									) : (
										<Button size="sm" className="w-full gap-1.5" onClick={() => setConnected(true)} disabled={!url.trim()}>
											<Zap className="size-3.5" aria-hidden />
											Connect
										</Button>
									)}
								</div>
							</div>

							<p className="font-mono text-[10px] text-muted-foreground/70">Socket.IO negotiates Engine.IO under the hood. Use <code className="rounded bg-muted px-1">/</code> for default.</p>
						</div>

						<div className="flex flex-col gap-2.5 p-3">
							<div className="flex items-center gap-2">
								<Cable className="size-3.5 text-muted-foreground" aria-hidden />
								<h3 className="text-xs font-semibold tracking-tight">Emit</h3>
								<span className="text-[11px] text-muted-foreground">— send one event</span>
							</div>

							<div className="grid grid-cols-[1.2fr_1fr] gap-2">
								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Event</span>
									<Input value={eventName} onChange={(e) => setEventName(e.target.value)} placeholder="chat:message" spellCheck={false} className="font-mono text-xs" />
								</label>
								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Namespace</span>
									<Input value={namespace} onChange={(e) => setNamespace(e.target.value)} placeholder="/" spellCheck={false} className="font-mono text-xs" />
								</label>
							</div>

							<label className="flex flex-col gap-1">
								<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Data (JSON)</span>
								<Textarea value={eventData} onChange={(e) => setEventData(e.target.value)} rows={3} spellCheck={false} className="font-mono text-xs" placeholder='{"text":"hello"}' />
								{eventData.trim() !== "" && !isJson(eventData) && <span className="font-mono text-[10px] text-status-warn">Not JSON — will be sent as raw string.</span>}
							</label>

							<div className="flex items-center gap-2">
								<span className="flex-1" />
								<Button size="sm" onClick={() => void emit()} disabled={busy}>{busy ? "Emitting…" : "Emit →"}</Button>
							</div>

							{error && (
								<Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>
							)}
							{ok && <p className="font-mono text-xs text-status-ok">✓ {ok}</p>}
						</div>
					</div>
				</div>

				<div className="flex min-h-0 flex-1 flex-col">
					<div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/20 px-3 py-2">
						<h4 className="flex items-center gap-1.5 text-xs font-semibold">
							<Radio className="size-3.5 text-muted-foreground" aria-hidden />
							Ledger<span className="font-mono text-[11px] font-normal text-muted-foreground">· {filtered.length} frames · {connected ? "live" : "idle"}</span>
						</h4>
						<span className="hidden h-3 w-px bg-border sm:block" aria-hidden />
						<div className="flex items-center gap-1 font-mono text-[10px] text-muted-foreground">
							<Filter className="size-3" aria-hidden />
							<Input value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="filter event or payload" className="h-6 w-44 font-mono text-xs" />
						</div>
						<span className="flex-1" />
						<Button variant="ghost" size="xs" onClick={() => setPaused((p) => !p)} className="gap-1">{paused ? <Play className="size-3.5" /> : <Pause className="size-3.5" />}{paused ? "Resume" : "Pause"}</Button>
						<Button variant="ghost" size="xs" onClick={() => setFrames([])} className="gap-1 text-muted-foreground"><Trash2 className="size-3.5" />Clear</Button>
					</div>

					<div ref={logRef} className="min-h-0 flex-1 overflow-y-auto p-2" aria-label="Socket.IO event ledger" style={{ backgroundImage: `linear-gradient(to right, color-mix(in srgb, var(--border) 6%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in srgb, var(--border) 6%, transparent) 1px, transparent 1px)`, backgroundSize: "24px 24px" }}>
						{filtered.length === 0 ? (
							<div className="flex h-full min-h-[160px] flex-col items-center justify-center gap-2 rounded border border-dashed border-border bg-card/60 px-4 py-10 text-center">
								<Cable className="size-5 text-muted-foreground/50" aria-hidden />
								<p className="max-w-[44ch] text-balance text-sm font-medium">{connected ? "Listening for events… emit to see the ledger move." : "Connect to start the ledger."}</p>
								<p className="max-w-[52ch] text-balance font-mono text-xs text-muted-foreground">Namespace isolates channels; events are the verb. Payload is typically JSON. Try <code className="rounded bg-muted px-1">/</code> then <code className="rounded bg-muted px-1">chat:message</code>.</p>
								<div className="mt-2 flex flex-wrap justify-center gap-1.5 font-mono text-[11px]">
									<button type="button" onClick={() => setEventName("chat:message")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">chat:message</button>
									<button type="button" onClick={() => setEventName("presence:join")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">presence:join</button>
									<button type="button" onClick={() => setNamespace("/chat")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">/chat</button>
								</div>
							</div>
						) : (
							<ul className="flex flex-col gap-1.5">
								{filtered.map((f, idx) => {
									if (f.type !== "message") {
										return (
											<li key={`${f.timestamp}-${idx}`} className="flex items-center gap-2 rounded border border-dashed border-border bg-muted/20 px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground">
												<span className="size-1.5 rounded-full bg-status-info" aria-hidden />
												<span>{f.raw ?? f.type}</span>
												<span className="ml-auto tabular-nums">{formatTime(f.timestamp)}</span>
											</li>
										);
									}
									const payloadStr = isString(f.data) ? f.data : JSON.stringify(f.data ?? "", null, 2);
									return (
										<li key={`${f.timestamp}-${idx}-${f.event}`} className="group flex items-stretch gap-0 overflow-hidden rounded border border-border bg-card text-xs shadow-[0_1px_0_color-mix(in_srgb,var(--border)_50%,transparent)]">
											<span aria-hidden className="w-[3px] shrink-0 bg-status-redirect" />
											<div className="flex min-w-0 flex-1 flex-col gap-1 px-2.5 py-1.5">
												<div className="flex flex-wrap items-center gap-1.5">
													<span className="tabular-nums font-mono text-[10px] text-muted-foreground">{formatTime(f.timestamp)}</span>
													<span className="text-border-strong">·</span>
													<span className="rounded border border-status-redirect/20 bg-status-redirect/10 px-1 py-0.5 font-mono text-[10px] font-medium text-status-redirect">{f.event ?? "event"}</span>
													{f.namespace && f.namespace !== "/" && <span className="rounded border bg-muted/40 px-1 py-0.5 font-mono text-[10px] text-muted-foreground">ns:{f.namespace}</span>}
													<span className="ml-auto flex items-center gap-1">
														<button type="button" onClick={() => void copyText(payloadStr)} className="inline-flex size-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:border-border hover:bg-muted/60 hover:text-foreground" aria-label="Copy payload" title="Copy payload"><Copy className="size-3" aria-hidden /></button>
													</span>
												</div>
												<pre className="max-h-28 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/40 px-2 py-1.5 font-mono text-[11px] leading-relaxed">{payloadStr}</pre>
											</div>
										</li>
									);
								})}
							</ul>
						)}
					</div>

					<div className="flex items-center gap-2 border-t border-border bg-muted/10 px-3 py-1.5 font-mono text-[10px] text-muted-foreground">
						<span>{url || "—"} {namespace ? `· ns:${namespace}` : ""}</span>
						<span className="ml-auto hidden sm:inline">Ledger is append-only · <code className="rounded bg-muted px-1">⌘K</code> to jump</span>
					</div>
				</div>
			</div>
		</section>
	);
}
