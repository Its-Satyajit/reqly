import { useEffect, useState } from "react";
import { Check, Copy, FileText, Pencil, Play, Plus, Square, Trash2 } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { SplitView, ViewShell } from "../../components/shell/ViewLayout";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { cn } from "#lib/utils";
import { methodTintClass } from "#lib/status";
import { MOCK_METHOD_OPTIONS } from "#lib/mock";
import { notifyError, notifySuccess } from "#lib/notify";
import { useMockStore } from "#stores/useMockStore";
import { copyText } from "#lib/response";

/** Parses a numeric input value, falling back when the text isn't a number. */
function inputInt(value: string, fallback: number): number {
	const parsed = Number(value);
	return Number.isFinite(parsed) ? parsed : fallback;
}

const numberInput =
	"rounded-md border border-border bg-transparent px-2 py-1 text-xs font-mono w-20";

function FieldLabel({ children }: { children: string }) {
	return (
		<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
			{children}
		</p>
	);
}

/** RouteEditor is the expanded per-route editing form. */
function RouteEditor({
	route,
	index,
	onUpdate,
	onRemove,
}: {
	route: ReturnType<typeof useMockStore.getState>["routes"][number];
	index: number;
	onUpdate: ReturnType<typeof useMockStore.getState>["updateRoute"];
	onRemove: ReturnType<typeof useMockStore.getState>["removeRoute"];
}) {
	return (
		<div className="flex flex-col gap-2 border-t border-border/50 px-3 py-2">
			<div className="flex items-center gap-1.5">
				<input
					type="checkbox"
					checked={route.enabled}
					onChange={(e) => onUpdate(index, { enabled: e.target.checked })}
					title={route.enabled ? "Enabled" : "Disabled"}
					aria-label={`Route ${route.path} enabled`}
					className="size-3.5 shrink-0 accent-(--primary)"
				/>
				<select
					value={route.method}
					onChange={(e) => onUpdate(index, { method: e.target.value })}
					aria-label={`Route ${route.path} method`}
					className="w-20 rounded-md border border-border bg-transparent px-1 py-1 text-xs"
				>
					{MOCK_METHOD_OPTIONS.map((o) => (
						<option key={o.value} value={o.value}>
							{o.label}
						</option>
					))}
				</select>
				<Input
					value={route.path}
					onChange={(e) => onUpdate(index, { path: e.target.value })}
					placeholder="/path"
					spellCheck={false}
					aria-label={`Route ${route.path} path`}
					className="min-w-0 flex-1 font-mono text-xs"
				/>
				<input
					type="number"
					value={route.status || 200}
					onChange={(e) =>
						onUpdate(index, { status: inputInt(e.target.value, route.status || 200) })
					}
					aria-label={`Route ${route.path} status`}
					className={cn(numberInput)}
					min={100}
					max={599}
				/>
				<Button
					variant="ghost"
					size="sm"
					aria-label={`Remove route ${route.path}`}
					onClick={() => onRemove(index)}
				>
					<Trash2 />
				</Button>
			</div>
			<Textarea
				value={route.body}
				onChange={(e) => onUpdate(index, { body: e.target.value })}
				rows={3}
				spellCheck={false}
				aria-label={`Route ${route.path} response body`}
				placeholder='Response body, e.g. {"ok":true}'
				className="resize-y font-mono text-xs"
			/>
			<Textarea
				value={route.headerLines.join("\n")}
				onChange={(e) => onUpdate(index, { headerLines: e.target.value.split("\n") })}
				rows={2}
				spellCheck={false}
				aria-label={`Route ${route.path} response headers`}
				placeholder={"Content-Type: application/json\nX-Mock: true"}
				className="resize-y font-mono text-xs"
			/>
		</div>
	);
}

/** MocksView is the G-17.4.5 mock-server surface: a servers panel with quick
 * templates beside a running header, endpoints table, and per-route Edit
 * expansion. Single-server (bridge seam); hit counters / live logs land with
 * a backend reporting seam. */
export function MocksView() {
	const specPath = useMockStore((s) => s.specPath);
	const port = useMockStore((s) => s.port);
	const delayMs = useMockStore((s) => s.delayMs);
	const failEvery = useMockStore((s) => s.failEvery);
	const routes = useMockStore((s) => s.routes);
	const status = useMockStore((s) => s.status);
	const busy = useMockStore((s) => s.busy);
	const error = useMockStore((s) => s.error);
	const setSpecPath = useMockStore((s) => s.setSpecPath);
	const setPort = useMockStore((s) => s.setPort);
	const setDelayMs = useMockStore((s) => s.setDelayMs);
	const setFailEvery = useMockStore((s) => s.setFailEvery);
	const updateRoute = useMockStore((s) => s.updateRoute);
	const addRoute = useMockStore((s) => s.addRoute);
	const removeRoute = useMockStore((s) => s.removeRoute);
	const start = useMockStore((s) => s.start);
	const stop = useMockStore((s) => s.stop);
	const [editingId, setEditingId] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);

	// Re-sync the panel when the backend restarts outside this view.
	useEffect(() => {
		void useMockStore.getState().refreshStatus();
	}, []);

	const copyBaseUrl = (): void => {
		if (!status.url) return;
		void copyText(status.url).then((ok) => {
			if (!ok) {
				notifyError("Copy failed", "Clipboard access was denied.");
				return;
			}
			setCopied(true);
			setTimeout(() => setCopied(false), 1500);
		});
	};

	return (
		<ViewShell label="Mock server">
			<SplitView
				asideLabel="Mock servers"
				aside={
					<>
				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<div className="flex items-center justify-between">
						<FieldLabel>Servers</FieldLabel>
						<Button
							size="icon"
							variant="ghost"
							aria-label="Add route"
							onClick={addRoute}
						>
							<Plus className="size-3.5" />
						</Button>
					</div>
					<button
						type="button"
						className={cn(
							"flex flex-col gap-0.5 rounded-lg border px-2.5 py-2 text-left",
							status.running
								? "border-primary/50 bg-primary/5"
								: "border-border bg-muted/20",
						)}
					>
						<span className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
							<span
								aria-hidden
								className={cn(
									"size-2 rounded-full",
									status.running ? "bg-status-ok" : "bg-muted-foreground/40",
								)}
							/>
							Mock server
						</span>
						<span className="truncate font-data text-xs text-muted-foreground">
							{status.running ? status.url : `localhost:${port} · stopped`}
						</span>
					</button>
				</div>
				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<FieldLabel>Quick templates</FieldLabel>
					<button
						type="button"
						className="flex items-center gap-2 rounded-md px-1.5 py-1 text-left text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
						onClick={() => {
							setSpecPath("");
							notifySuccess("Blank server", "Spec cleared — manual routes apply.");
						}}
					>
						<FileText className="size-3.5 shrink-0" aria-hidden />
						Blank server
					</button>
					<button
						type="button"
						className="flex items-center gap-2 rounded-md px-1.5 py-1 text-left text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
						onClick={() => notifySuccess("From OpenAPI spec", "Set the spec path below, then Start.")}
					>
						<FileText className="size-3.5 shrink-0" aria-hidden />
						From OpenAPI spec
					</button>
				</div>
				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<FieldLabel>Port</FieldLabel>
					<input
						type="number"
						value={port}
						onChange={(e) => setPort(inputInt(e.target.value, port))}
						className={numberInput}
						min={1}
						max={65535}
						aria-label="Port"
					/>
					<FieldLabel>Delay (ms)</FieldLabel>
					<input
						type="number"
						value={delayMs}
						onChange={(e) => setDelayMs(inputInt(e.target.value, delayMs))}
						className={numberInput}
						min={0}
						aria-label="Delay (ms)"
					/>
					<FieldLabel>Fail every Nth</FieldLabel>
					<input
						type="number"
						value={failEvery}
						onChange={(e) => setFailEvery(inputInt(e.target.value, failEvery))}
						className={numberInput}
						min={0}
						aria-label="Fail every Nth request"
					/>
					<FieldLabel>OpenAPI spec (optional)</FieldLabel>
					<Input
						value={specPath}
						onChange={(e) => setSpecPath(e.target.value)}
						placeholder="specs/pets.yaml"
						spellCheck={false}
						className="font-mono text-xs"
						aria-label="OpenAPI spec path"
					/>
				</div>
					</>
			}
		>
			<section className="flex min-w-0 flex-1 flex-col gap-3 overflow-y-auto">
				<div className="flex flex-wrap items-center gap-2">
					<h2 className="flex items-center gap-2 text-base font-semibold text-foreground">
						<span
							aria-hidden
							className={cn(
								"size-2.5 rounded-full",
								status.running ? "bg-status-ok" : "bg-muted-foreground/40",
							)}
						/>
						Mock server
					</h2>
					{status.running ? (
						<span className="rounded-full border border-status-ok/40 px-2 py-0.5 font-data text-2xs text-status-ok">
							running
						</span>
					) : null}
					<div className="ml-auto flex items-center gap-2">
						{status.running ? (
							<>
								<Button variant="outline" size="sm" disabled={busy} onClick={() => void stop()}>
									{busy ? <Spinner data-icon="inline-start" /> : <Square data-icon="inline-start" />}
									Stop
								</Button>
								<Button variant="outline" size="sm" onClick={copyBaseUrl}>
									{copied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
									{copied ? "Copied" : "Copy base URL"}
								</Button>
							</>
						) : (
							<Button variant="destructive" size="sm" disabled={busy} onClick={() => void start()}>
								{busy ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
								Start
							</Button>
						)}
					</div>
				</div>
				{status.running && status.url ? (
					<p className="font-data text-xs text-muted-foreground">
						{status.url} · delay {delayMs}ms
						{failEvery > 0 ? ` · fail every ${failEvery}th` : ""}
					</p>
				) : null}

				{error ? (
					<Alert variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				) : null}

				<div className="flex items-center justify-between">
					<h3 className="text-sm font-semibold text-foreground">Endpoints</h3>
					<Button variant="outline" size="sm" onClick={addRoute}>
						<Plus data-icon="inline-start" />
						Add route
					</Button>
				</div>

				<div className="overflow-hidden rounded-xl border border-border bg-card">
					<table className="w-full border-separate border-spacing-0 text-left text-xs">
						<thead>
							<tr>
								{["Method", "Path", "Status", "Delay", ""].map((h) => (
									<th
										key={h}
										className="border-b border-border px-3 py-1.5 font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground"
									>
										{h}
									</th>
								))}
							</tr>
						</thead>
						<tbody>
							{routes.length === 0 ? (
								<tr>
									<td colSpan={5} className="px-3 py-4 text-center text-muted-foreground">
										No routes — Add route or point at an OpenAPI spec, then Start.
									</td>
								</tr>
							) : (
								routes.map((route) => (
									<tr key={route.id ?? `${route.method}-${route.path}`}>
										<td className="border-b border-border/40 px-3 py-1.5">
											<span
												className={cn(
													"rounded-full border border-border bg-muted/40 px-1.5 py-px font-data text-2xs font-semibold uppercase",
													methodTintClass(route.method),
												)}
											>
												{route.method}
											</span>
										</td>
										<td className="border-b border-border/40 px-3 py-1.5 font-mono">
											{route.path || "—"}
										</td>
										<td className="border-b border-border/40 px-3 py-1.5 font-data tabular-nums">
											{route.status || 200}
										</td>
										<td className="border-b border-border/40 px-3 py-1.5 font-data tabular-nums text-muted-foreground">
											{delayMs} ms
										</td>
										<td className="border-b border-border/40 px-3 py-1.5 text-right">
											<Button
												variant="outline"
												size="xs"
												onClick={() =>
													setEditingId(editingId === route.id ? null : (route.id ?? null))
												}
											>
												<Pencil data-icon="inline-start" />
												Edit
											</Button>
										</td>
									</tr>
								))
							)}
						</tbody>
					</table>
					{routes.map((route, i) =>
						editingId != null && route.id === editingId ? (
							<RouteEditor
								key={`edit-${route.id}`}
								route={route}
								index={i}
								onUpdate={updateRoute}
								onRemove={removeRoute}
							/>
						) : null,
					)}
				</div>
			</section>
			</SplitView>
		</ViewShell>
	);
}
