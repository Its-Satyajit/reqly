import { useState } from "react";
import type { RunStep } from "#lib/collections";
import { ChevronRight, X } from "lucide-react";
import { cn } from "#lib/utils";
import { formatBytes } from "#lib/ui";
import { CodeMirrorEditor } from "../editors/CodeMirrorEditor";
import { Button } from "./ui/button";
import { Download } from "lucide-react";
import { useCollectionRunStore, useWorkspaceStore } from "#stores";

/** previewBody truncates a response body for the inline summary panel. */
const previewBody = (body: string): string =>
	body.length > 2000 ? `${body.slice(0, 2000)}\n… (${body.length - 2000} more bytes)` : body;

/** StepRow renders one streamed run step: name, live pass/fail, status code
 * and duration, with an expandable panel for tests, script logs and the
 * response summary. Clicking the step opens its request in a normal tab. */
function StepRow({ step, onOpenRequest }: { step: RunStep; onOpenRequest: (path: string) => void }) {
	const [expanded, setExpanded] = useState(false);
	const { response } = step;
	return (
		<div className="border-b border-border last:border-b-0">
			<div className="flex w-full items-center gap-2 px-3 py-1.5 text-xs">
				<button
					type="button"
					onClick={() => setExpanded((v) => !v)}
					className="shrink-0 text-muted-foreground/50 hover:text-foreground"
					title="Toggle details"
				>
					<ChevronRight
						className={cn("size-3 transition-transform", expanded && "rotate-90")}
						aria-hidden
					/>
				</button>
				<button
					type="button"
					onClick={() => onOpenRequest(step.requestPath)}
					className="flex min-w-0 flex-1 items-center gap-2 text-left"
					title="Open request"
				>
					<span
						className={cn(
							"size-2 shrink-0 rounded-full",
							step.passed ? "bg-status-ok" : "bg-status-error",
						)}
					/>
					<span className="truncate font-medium">{step.name}</span>
				</button>
				<span className="shrink-0 text-muted-foreground">
					{step.passed ? (
						<>
							{response?.statusCode ?? "—"}
							{response ? ` · ${response.durationMs}ms` : ""}
						</>
					) : (
						<span className="text-destructive">failed</span>
					)}
				</span>
			</div>
			{expanded && (
				<div className="space-y-2 px-4 pb-2">
					{step.requestError && <p className="text-xs text-destructive">{step.requestError}</p>}
					{step.tests.length > 0 && (
						<ul className="space-y-1">
							{step.tests.map((t) => (
								<li key={t.name} className="flex items-center gap-2 text-xs">
									<span
										className={cn(
											"size-1.5 rounded-full",
											t.passed ? "bg-status-ok" : "bg-status-error",
										)}
									/>
									<span className={cn(t.passed || "text-destructive")}>{t.name}</span>
									<span className="text-muted-foreground/60">{t.passed ? "passed" : "failed"}</span>
								</li>
							))}
						</ul>
					)}
					{step.logs.length > 0 && (
						<pre className="max-h-40 overflow-auto rounded-md bg-muted/50 p-2 font-mono text-[11px] leading-relaxed">
							{step.logs.join("\n")}
						</pre>
					)}
					{response && (
						<div className="rounded-md border border-border p-2">
							<div className="flex items-center gap-2 text-xs text-muted-foreground">
								<span>
									{response.statusCode} {response.statusText}
								</span>
								<span>·</span>
								<span>{response.durationMs}ms</span>
								<span>·</span>
								<span>{formatBytes(response.size)}</span>
							</div>
							<pre className="mt-1 max-h-40 overflow-auto font-mono text-[11px] leading-relaxed">
								{previewBody(response.body)}
							</pre>
						</div>
					)}
				</div>
			)}
		</div>
	);
}

/** RunView is the live collection-run surface: a summary header (status,
 * totals, duration, fail-fast toggle, cancel, dismissible unsaved-drafts hint)
 * above the ordered list of streamed steps. */
export function RunView() {
	const running = useCollectionRunStore((s) => s.running);
	const steps = useCollectionRunStore((s) => s.steps);
	const report = useCollectionRunStore((s) => s.report);
	const error = useCollectionRunStore((s) => s.error);
	const path = useCollectionRunStore((s) => s.path);
	const env = useCollectionRunStore((s) => s.env);
	const failFast = useCollectionRunStore((s) => s.failFast);
	const toggleFailFast = useCollectionRunStore((s) => s.toggleFailFast);
	const cancelRun = useCollectionRunStore((s) => s.cancelRun);
	const openRequest = useWorkspaceStore((s) => s.openRequest);
	const exportReport = useCollectionRunStore((s) => s.exportReport);
	const [dismissed, setDismissed] = useState(false);
	const [exporting, setExporting] = useState(false);
	const [preview, setPreview] = useState<{ path: string; content: string; language: "json" | "xml" } | null>(null);
	const [exportError, setExportError] = useState<string | null>(null);

	// G-16.1: serialize the finished run as JUnit XML or JSON, save it under
	// .reqly/exports/, and open the rendered text as an inline preview.
	// Promise-chained (not async/await) — the compiler cannot optimize
	// async closures declared in component scope.
	const doExport = (format: "json" | "junit"): void => {
		setExporting(true);
		setExportError(null);
		exportReport(format)
			.then((res) => {
				if (!res) {
					setExportError("Report export is not available for this run.");
					return;
				}
				setPreview({
					path: res.path,
					content: res.content,
					language: format === "json" ? "json" : "xml",
				});
			})
			.catch((failure: Error) => setExportError(failure.message))
			.finally(() => setExporting(false));
	};

	if (!path) {
		return (
			<div className="flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground">
				No run started — press the play button on a collection or folder to run it.
			</div>
		);
	}

	const status = running
		? "Running"
		: report
			? report.ok
				? "Passed"
				: "Failed"
			: error
				? "Failed"
				: "Idle";
	const statusDot = running ? "bg-primary" : status === "Passed" ? "bg-status-ok" : "bg-status-error";

	return (
		<div className="flex h-full min-h-0 flex-col">
			<div className="border-b border-border px-4 py-3">
				<div className="flex items-center justify-between gap-4">
					<div className="min-w-0">
						<h2 className="truncate text-sm font-semibold">{path}</h2>
						<div className="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
							<span className="inline-flex items-center gap-1.5">
								<span className={cn("size-1.5 rounded-full", statusDot)} />
								{status}
							</span>
							{env ? (
								<span className="rounded-full bg-muted px-1.5 py-0.5">{env}</span>
							) : (
								<span className="rounded-full bg-muted px-1.5 py-0.5">Default env</span>
							)}
							<span>
								{steps.length} of {report?.total ?? "…"} steps
							</span>
							{report && <span>· {report.durationMs}ms</span>}
						</div>
					</div>
					<div className="flex shrink-0 items-center gap-3">
						<label className="flex cursor-pointer items-center gap-1.5 text-xs text-muted-foreground">
							<input
								type="checkbox"
								checked={failFast}
								onChange={toggleFailFast}
								className="size-3"
							/>
							Fail fast
						</label>
						{report && (
							<Button
								variant="outline"
								size="sm"
								onClick={() => void doExport("junit")}
								disabled={exporting}
								title="Save the report as JUnit XML and preview it"
							>
								<Download data-icon="inline-start" />
								JUnit
							</Button>
						)}
						{report && (
							<Button
								variant="outline"
								size="sm"
								onClick={() => void doExport("json")}
								disabled={exporting}
								title="Save the report as JSON and preview it"
							>
								<Download data-icon="inline-start" />
								JSON
							</Button>
						)}
						<button
							type="button"
							onClick={() => void cancelRun()}
							disabled={!running}
							className="rounded-md border border-border px-2 py-1 text-xs disabled:cursor-not-allowed disabled:opacity-50"
						>
							Cancel
						</button>
					</div>
				</div>
				{exportError && (
					<p className="mt-2 text-[11px] text-status-error">{exportError}</p>
				)}
				{!dismissed && (
					<div className="mt-2 flex items-start justify-between gap-2 rounded-md bg-muted/50 px-2 py-1.5 text-[11px] text-muted-foreground">						<span>
							Runs execute the saved request files from disk — unsaved changes in request tabs are
							not included.
						</span>
						<button
							type="button"
							onClick={() => setDismissed(true)}
							className="shrink-0 text-muted-foreground/60 hover:text-foreground"
							title="Dismiss"
							aria-label="Dismiss report summary"
						>
							<X className="size-3.5" aria-hidden />
						</button>
					</div>
				)}
			</div>
			{preview && (
				<div className="flex min-h-0 flex-[2] flex-col border-b border-border">
					<div className="flex shrink-0 items-center justify-between gap-2 border-b border-border px-3 py-1.5 text-xs text-muted-foreground">
						<span className="truncate font-mono" title={preview.path}>
							{preview.path}
						</span>
						<Button variant="outline" size="sm" onClick={() => setPreview(null)}>
							Close preview
						</Button>
					</div>
					<CodeMirrorEditor
						value={preview.content}
						language={preview.language}
						readOnly
						className="min-h-0 flex-1 overflow-hidden"
					/>
				</div>
			)}
			<div className="min-h-0 flex-1 overflow-y-auto">
				{steps.length === 0 && !running && !report ? (
					<div className="p-8 text-center text-xs text-muted-foreground">Waiting for the run to start…</div>
				) : (
					<div>
						{steps.map((step, i) => (
							<StepRow
								key={`${step.requestPath}-${i}`}
								step={step}
								onOpenRequest={(p) => void openRequest(p)}
							/>
						))}
					</div>
				)}
				{error && <p className="px-4 py-2 text-xs text-destructive">{error}</p>}
			</div>
		</div>
	);
}