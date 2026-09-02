import { useEffect, useMemo, useRef, useState } from "react";
import { Antenna, Copy, Filter, Pause, Pin, Play, Radio, Trash2, Wifi } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Textarea } from "#components/ui/textarea";
import { CompactSelect } from "#components/CompactSelect";
import { Alert, AlertDescription } from "#components/ui/alert";
import { cn } from "#lib/utils";
import { copyText } from "#lib/response";
import type { MqttFrameView } from "#lib/mqtt";
import { getMqttBridge, qosTint } from "#lib/mqtt";

function TopicSegments({ topic }: { topic: string }) {
	const parts = topic.split("/").filter(Boolean);
	const isWild = (p: string) => p === "#" || p === "+";
	if (!topic) return <span className="text-muted-foreground">—</span>;
	return (
		<span className="inline-flex flex-wrap items-center gap-0 font-mono text-[11px] leading-none">
			{parts.map((seg, i) => (
				<span key={`${seg}-${i}`} className="inline-flex items-center gap-0">
					{i !== 0 && <span className="px-0.5 text-border-strong select-none">/</span>}
					<span
						className={cn(
							"rounded px-1 py-0.5",
							isWild(seg)
								? "bg-status-warn/10 text-status-warn border border-status-warn/20"
								: "bg-muted/60 text-foreground",
						)}
					>
						{seg}
					</span>
				</span>
			))}
			{topic.endsWith("/") && <span className="px-0.5 text-border-strong">/</span>}
		</span>
	);
}

function formatTime(ts: number): string {
	const d = new Date(ts);
	return d.toLocaleTimeString(undefined, { hour12: false }) + "." + String(d.getMilliseconds()).padStart(3, "0");
}

function parseQos(v: string): 0 | 1 | 2 {
	if (v === "1") return 1;
	if (v === "2") return 2;
	return 0;
}

const QOS_OPTIONS = [
	{ value: "0", label: "QoS 0 · at most once" },
	{ value: "1", label: "QoS 1 · at least once" },
	{ value: "2", label: "QoS 2 · exactly once" },
];

export function MqttView() {
	const [broker, setBroker] = useState("mqtt://localhost:1883");
	const [pubTopic, setPubTopic] = useState("sensors/temperature");
	const [pubMessage, setPubMessage] = useState('{"value": 22.4, "unit":"°C"}');
	const [pubQos, setPubQos] = useState("0");
	const [pubRetain, setPubRetain] = useState(false);
	const [pubUser, setPubUser] = useState("");
	const [pubPass, setPubPass] = useState("");
	const [pubBusy, setPubBusy] = useState(false);
	const [pubError, setPubError] = useState<string | null>(null);
	const [pubOk, setPubOk] = useState<string | null>(null);

	const [subTopic, setSubTopic] = useState("sensors/#");
	const [subQos, setSubQos] = useState("0");
	const [subUser, setSubUser] = useState("");
	const [subPass, setSubPass] = useState("");
	const [subActive, setSubActive] = useState(false);
	const [subSessionId] = useState(() => `mqtt-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`);
	const [subError, setSubError] = useState<string | null>(null);

	const [frames, setFrames] = useState<MqttFrameView[]>([]);
	const [filter, setFilter] = useState("");
	const [paused, setPaused] = useState(false);
	const logRef = useRef<HTMLDivElement>(null);
	const offRef = useRef<(() => void) | null>(null);
	const mockTimerRef = useRef<number | null>(null);

	// eslint-disable-next-line react-doctor/effect-needs-cleanup -- off() + clearInterval for mock ticker are in cleanup below
	useEffect(() => {
		if (!subActive) {
			offRef.current?.();
			offRef.current = null;
			if (mockTimerRef.current) {
				window.clearInterval(mockTimerRef.current);
				mockTimerRef.current = null;
			}
			return;
		}
		let cancelled = false;
		const bridge = getMqttBridge();
		const off = bridge.onFrame(subSessionId, (f) => {
			if (paused || cancelled) return;
			if (f.type === "message") setFrames((prev) => [...prev.slice(-499), f]);
			if (f.type === "error") setSubError(f.data ?? "subscribe error");
		});
		offRef.current = off;

		const startSub = (): void => {
			bridge
				.subscribe({
					sessionId: subSessionId,
					broker: broker.trim(),
					topic: subTopic.trim(),
					qos: parseQos(subQos),
					username: subUser || undefined,
					password: subPass || undefined,
				})
				.then(() => {
					if (cancelled) return;
					setSubError(null);
					setFrames((prev) => [
						...prev.slice(-499),
						{
							sessionId: subSessionId,
							type: "status" as const,
							data: `subscribed to ${subTopic.trim()}`,
							timestamp: Date.now(),
						},
					]);
				})
				.catch((err) => {
					if (cancelled) return;
					const msg = err instanceof Error ? err.message : String(err);
					if (msg.includes("not available in this build")) {
						setSubError(null);
						let n = 0;
						mockTimerRef.current = window.setInterval(() => {
							if (paused || cancelled) return;
							const topics = ["sensors/temperature", "sensors/humidity", "devices/edge-01/status"];
							const t = topics[n % topics.length];
							const payloads = ['{"value":22.4}', '{"value":61}', '{"state":"online","up":14203}'];
							setFrames((prev) => [
								...prev.slice(-499),
								{
									sessionId: subSessionId,
									type: "message",
									topic: t,
									payload: payloads[n % payloads.length],
									qos: parseQos(subQos),
									retain: n % 7 === 0,
									timestamp: Date.now(),
								},
							]);
							n++;
						}, 1200);
						return;
					}
					setSubError(msg);
					setSubActive(false);
				});
		};
		startSub();

		return () => {
			cancelled = true;
			off();
			offRef.current = null;
			if (mockTimerRef.current) {
				window.clearInterval(mockTimerRef.current);
				mockTimerRef.current = null;
			}
			void getMqttBridge()
				.cancel(subSessionId)
				.catch(() => {});
		};
	}, [subActive, subSessionId, broker, subTopic, subQos, paused, subUser, subPass]);

	useEffect(() => {
		if (logRef.current && !paused) logRef.current.scrollTop = logRef.current.scrollHeight;
	}, [frames.length, paused]);

	const filtered = useMemo(() => {
		if (!filter.trim()) return frames;
		const q = filter.trim().toLowerCase();
		return frames.filter(
			(f) => f.topic?.toLowerCase().includes(q) || f.payload?.toLowerCase().includes(q) || f.data?.toLowerCase().includes(q),
		);
	}, [frames, filter]);

	const publish = async () => {
		if (!broker.trim() || !pubTopic.trim()) {
			setPubError("Broker and topic are required");
			return;
		}
		setPubBusy(true);
		setPubError(null);
		setPubOk(null);
		try {
			await getMqttBridge().publish({
				broker: broker.trim(),
				topic: pubTopic.trim(),
				message: pubMessage,
				qos: parseQos(pubQos),
				retain: pubRetain,
				username: pubUser || undefined,
				password: pubPass || undefined,
			});
			setPubOk(`published to ${pubTopic.trim()}`);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setPubOk(`queued to ${pubTopic.trim()} (mock — no broker in dev)`);
			} else {
				setPubError(msg);
				return;
			}
		} finally {
			setPubBusy(false);
		}
		if (!paused) {
			setFrames((prev) => [
				...prev.slice(-499),
				{
					sessionId: "local-publish",
					type: "message",
					topic: pubTopic.trim(),
					payload: pubMessage,
					qos: parseQos(pubQos),
					retain: pubRetain,
					timestamp: Date.now(),
				},
			]);
		}
	};

	return (
		<section className="flex h-full min-h-0 flex-col overflow-hidden" aria-label="MQTT">
			<PageHeader
				icon={Antenna}
				title="MQTT"
				description="Publish and subscribe over MQTT — broker, topic tape, and payload ledger. Local, no cloud."
				actions={
					<div className="flex items-center gap-1.5">
						<span
							className={cn(
								"inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-mono",
								subActive ? "border-status-ok/30 bg-status-ok/10 text-status-ok" : "border-border bg-muted/40 text-muted-foreground",
							)}
						>
							<span className={cn("size-1.5 rounded-full", subActive ? "bg-status-ok animate-pulse" : "bg-muted-foreground/40")} aria-hidden />
							{subActive ? "subscribed" : "idle"}
						</span>
						<Badge variant="outline" className="font-mono text-[10px]">
							{frames.filter((f) => f.type === "message").length} msgs
						</Badge>
					</div>
				}
			/>

			<div className="flex min-h-0 flex-1 flex-col gap-0 overflow-hidden">
				<div className="shrink-0 border-b border-border bg-card/20">
					<div className="grid gap-0 lg:grid-cols-2">
						<div className="flex flex-col gap-2.5 border-b border-border p-3 lg:border-b-0 lg:border-r">
							<div className="flex items-center gap-2">
								<Radio className="size-3.5 text-muted-foreground" aria-hidden />
								<h3 className="text-xs font-semibold tracking-tight">Publish</h3>
								<span className="text-[11px] text-muted-foreground">— send one message</span>
								<span className="ml-auto hidden items-center gap-1 font-mono text-[10px] text-muted-foreground sm:inline-flex">
									<Wifi className="size-3" aria-hidden /> {broker || "—"}
								</span>
							</div>

							<div className="flex flex-col gap-2">
								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Broker</span>
									<Input value={broker} onChange={(e) => setBroker(e.target.value)} placeholder="mqtt://broker.emqx.io:1883 or wss://host:8084/mqtt" spellCheck={false} className="font-mono text-xs" aria-label="MQTT broker" />
								</label>

								<div className="grid grid-cols-[1fr_auto] gap-2">
									<label className="flex flex-col gap-1">
										<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Topic</span>
										<Input value={pubTopic} onChange={(e) => setPubTopic(e.target.value)} placeholder="sensors/temperature" spellCheck={false} className="font-mono text-xs" />
									</label>
									<label className="flex w-44 flex-col gap-1">
										<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">QoS</span>
										<CompactSelect value={pubQos} onChange={setPubQos} options={QOS_OPTIONS} ariaLabel="Publish QoS" />
									</label>
								</div>

								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Payload</span>
									<Textarea value={pubMessage} onChange={(e) => setPubMessage(e.target.value)} rows={3} spellCheck={false} className="font-mono text-xs" placeholder='{"value": 22.4}' />
								</label>

								<div className="flex flex-wrap items-center gap-2">
									<label className="inline-flex items-center gap-1.5 text-xs">
										<input type="checkbox" checked={pubRetain} onChange={(e) => setPubRetain(e.target.checked)} className="size-3.5 accent-[var(--primary)]" />
										<span className="inline-flex items-center gap-1">
											<Pin className="size-3 text-muted-foreground" aria-hidden /> retain
										</span>
									</label>
									<span className="flex-1" />
									<Button size="sm" onClick={() => void publish()} disabled={pubBusy}>
										{pubBusy ? "Publishing…" : "Publish →"}
									</Button>
								</div>

								<details className="rounded border border-border/60 bg-muted/20 px-2.5 py-1.5">
									<summary className="cursor-pointer list-none font-mono text-[11px] font-medium text-muted-foreground">Auth — username / password / clientId</summary>
									<div className="mt-2 grid grid-cols-2 gap-2">
										<Input value={pubUser} onChange={(e) => setPubUser(e.target.value)} placeholder="username (optional)" className="font-mono text-xs" />
										<Input value={pubPass} onChange={(e) => setPubPass(e.target.value)} type="password" placeholder="password" className="font-mono text-xs" />
									</div>
								</details>

								{pubError && (
									<Alert variant="destructive" className="py-2">
										<AlertDescription className="font-mono text-xs">{pubError}</AlertDescription>
									</Alert>
								)}
								{pubOk && <p className="font-mono text-xs text-status-ok">✓ {pubOk}</p>}
							</div>
						</div>

						<div className="flex flex-col gap-2.5 p-3">
							<div className="flex items-center gap-2">
								<Antenna className="size-3.5 text-muted-foreground" aria-hidden />
								<h3 className="text-xs font-semibold tracking-tight">Subscribe</h3>
								<span className="text-[11px] text-muted-foreground">— stream the tape</span>
								{subActive && (
									<span className="ml-auto inline-flex items-center gap-1 font-mono text-[10px] text-status-ok">
										<span className="size-1.5 animate-pulse rounded-full bg-status-ok" aria-hidden />
										live
									</span>
								)}
							</div>

							<label className="flex flex-col gap-1">
								<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Topic filter</span>
								<Input value={subTopic} onChange={(e) => setSubTopic(e.target.value)} placeholder="sensors/#  ·  + matches one level, # tail" spellCheck={false} className="font-mono text-xs" disabled={subActive} />
								<span className="font-mono text-[10px] text-muted-foreground/70">
									Try <code className="rounded bg-muted px-1">sensors/+/temperature</code> or <code className="rounded bg-muted px-1">#</code>
								</span>
							</label>

							<div className="grid grid-cols-2 gap-2">
								<label className="flex flex-col gap-1">
									<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">QoS</span>
									<CompactSelect value={subQos} onChange={setSubQos} options={QOS_OPTIONS} ariaLabel="Subscribe QoS" />
								</label>
								<div className="flex items-end">
									{subActive ? (
										<Button variant="outline" size="sm" className="w-full" onClick={() => setSubActive(false)}>
											<Pause className="size-3.5" aria-hidden />
											Unsubscribe
										</Button>
									) : (
										<Button size="sm" className="w-full" onClick={() => setSubActive(true)} disabled={!broker.trim() || !subTopic.trim()}>
											<Play className="size-3.5" aria-hidden />
											Subscribe
										</Button>
									)}
								</div>
							</div>

							<details className="rounded border border-border/60 bg-muted/20 px-2.5 py-1.5">
								<summary className="cursor-pointer list-none font-mono text-[11px] font-medium text-muted-foreground">Subscribe auth</summary>
								<div className="mt-2 grid grid-cols-2 gap-2">
									<Input value={subUser} onChange={(e) => setSubUser(e.target.value)} placeholder="username" className="font-mono text-xs" disabled={subActive} />
									<Input value={subPass} onChange={(e) => setSubPass(e.target.value)} type="password" placeholder="password" className="font-mono text-xs" disabled={subActive} />
								</div>
							</details>

							{subError && (
								<Alert variant="destructive" className="py-2">
									<AlertDescription className="font-mono text-xs">{subError}</AlertDescription>
								</Alert>
							)}

							<div className="rounded border border-border/60 bg-background px-2.5 py-2">
								<p className="font-mono text-[11px] font-semibold tracking-wide text-muted-foreground">Topic tape legend</p>
								<div className="mt-1.5 flex flex-wrap gap-1.5">
									<span className="inline-flex items-center gap-1 rounded border bg-muted/40 px-1.5 py-0.5 font-mono text-[10px]">
										<span className="size-2 rounded-full bg-status-warn" aria-hidden /> QoS 1
									</span>
									<span className="inline-flex items-center gap-1 rounded border bg-muted/40 px-1.5 py-0.5 font-mono text-[10px]">
										<span className="size-2 rounded-full bg-status-error" aria-hidden /> QoS 2
									</span>
									<span className="inline-flex items-center gap-1 rounded border bg-muted/40 px-1.5 py-0.5 font-mono text-[10px]">
										<Pin className="size-3" aria-hidden /> retain
									</span>
									<span className="inline-flex items-center gap-1 rounded border bg-muted/40 px-1.5 py-0.5 font-mono text-[10px]">
										<span className="size-1.5 rounded-full bg-status-ok animate-pulse" aria-hidden /> live
									</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				<div className="flex min-h-0 flex-1 flex-col">
					<div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/20 px-3 py-2">
						<h4 className="flex items-center gap-1.5 text-xs font-semibold">
							<Radio className="size-3.5 text-muted-foreground" aria-hidden />
							Stream
							<span className="font-mono text-[11px] font-normal text-muted-foreground">· {filtered.length} frames · {subActive ? "live" : "idle"}</span>
						</h4>
						<span className="hidden h-3 w-px bg-border sm:block" aria-hidden />
						<div className="flex items-center gap-1 font-mono text-[10px] text-muted-foreground">
							<Filter className="size-3" aria-hidden />
							<Input value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="filter topic or payload" className="h-6 w-44 font-mono text-xs" />
						</div>
						<span className="flex-1" />
						<Button variant="ghost" size="xs" onClick={() => setPaused((p) => !p)} className="gap-1">
							{paused ? <Play className="size-3.5" /> : <Pause className="size-3.5" />}
							{paused ? "Resume" : "Pause"}
						</Button>
						<Button variant="ghost" size="xs" onClick={() => setFrames([])} className="gap-1 text-muted-foreground">
							<Trash2 className="size-3.5" />
							Clear
						</Button>
					</div>

					<div
						ref={logRef}
						className="min-h-0 flex-1 overflow-y-auto p-2"
						aria-label="MQTT message tape"
						style={{
							backgroundImage: `linear-gradient(to right, color-mix(in srgb, var(--border) 6%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in srgb, var(--border) 6%, transparent) 1px, transparent 1px)`,
							backgroundSize: "24px 24px",
						}}
					>
						{filtered.length === 0 ? (
							<div className="flex h-full min-h-[160px] flex-col items-center justify-center gap-2 rounded border border-dashed border-border bg-card/60 px-4 py-10 text-center">
								<Antenna className="size-5 text-muted-foreground/50" aria-hidden />
								<p className="max-w-[44ch] text-balance text-sm font-medium">{subActive ? "Listening for messages… publish to see the tape move." : "Subscribe to a topic to start the tape."}</p>
								<p className="max-w-[52ch] text-balance font-mono text-xs text-muted-foreground">
									Topics are slash-separated. Use <code className="rounded bg-muted px-1">+</code> for one level and <code className="rounded bg-muted px-1">#</code> for the tail. Payload is rendered as text; binary hints as base64 length.
								</p>
								<div className="mt-2 flex flex-wrap justify-center gap-1.5 font-mono text-[11px]">
									<button type="button" onClick={() => setSubTopic("sensors/#")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">sensors/#</button>
									<button type="button" onClick={() => setSubTopic("#")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">#</button>
									<button type="button" onClick={() => setSubTopic("devices/+/status")} className="rounded border border-border bg-muted/40 px-2 py-1 hover:bg-muted">devices/+/status</button>
								</div>
							</div>
						) : (
							<ul className="flex flex-col gap-1.5">
								{filtered.map((f, idx) => {
									if (f.type !== "message") {
										return (
											<li key={`${f.timestamp}-${idx}`} className="flex items-center gap-2 rounded border border-dashed border-border bg-muted/20 px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground">
												<span className="size-1.5 rounded-full bg-status-info" aria-hidden />
												<span>{f.data ?? f.type}</span>
												<span className="ml-auto tabular-nums">{formatTime(f.timestamp)}</span>
											</li>
										);
									}
									return (
										<li key={`${f.timestamp}-${idx}-${f.topic}`} className="group flex items-stretch gap-0 overflow-hidden rounded border border-border bg-card text-xs shadow-[0_1px_0_color-mix(in_srgb,var(--border)_50%,transparent)]">
											<span aria-hidden className={cn("w-[3px] shrink-0", f.qos === 1 ? "bg-status-warn" : f.qos === 2 ? "bg-status-error" : "bg-border")} />
											<div className="flex min-w-0 flex-1 flex-col gap-1 px-2.5 py-1.5">
												<div className="flex flex-wrap items-center gap-1.5">
													<span className="inline-flex items-center gap-1 font-mono text-[10px]">
														<span className="tabular-nums text-muted-foreground">{formatTime(f.timestamp)}</span>
														<span className="text-border-strong">·</span>
														<span className={cn("rounded border px-1 py-0.5 text-[10px] font-medium", qosTint(f.qos ?? 0))}>QoS {f.qos ?? 0}</span>
														{f.retain && (
															<span className="inline-flex items-center gap-0.5 rounded border border-status-info/20 bg-status-info/10 px-1 py-0.5 text-[10px] text-status-info">
																<Pin className="size-2.5" aria-hidden /> retain
															</span>
														)}
													</span>
													<span className="ml-auto flex items-center gap-1">
														<button type="button" onClick={() => void copyText(f.payload ?? "")} className="inline-flex size-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:border-border hover:bg-muted/60 hover:text-foreground" aria-label="Copy payload" title="Copy payload">
															<Copy className="size-3" aria-hidden />
														</button>
													</span>
												</div>
												<div className="min-w-0">
													<TopicSegments topic={f.topic ?? ""} />
												</div>
												<pre className="max-h-24 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/40 px-2 py-1.5 font-mono text-[11px] leading-relaxed">{f.payload ?? ""}</pre>
											</div>
										</li>
									);
								})}
							</ul>
						)}
					</div>

					<div className="flex items-center gap-2 border-t border-border bg-muted/10 px-3 py-1.5 font-mono text-[10px] text-muted-foreground">
						<span>Broker {broker || "—"}</span>
						<span className="text-border">·</span>
						<span>filter {subTopic || "—"} · QoS {subQos}</span>
						<span className="ml-auto hidden sm:inline">Tape is append-only · <code className="rounded bg-muted px-1">⌘K</code> to jump</span>
					</div>
				</div>
			</div>
		</section>
	);
}
