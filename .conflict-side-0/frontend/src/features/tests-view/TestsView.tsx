import { useEffect, useState } from "react";
import { Download, FilePlus, Play, Pencil } from "lucide-react";
import { Button } from "#components/ui/button";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import type { TestFileRef, TestRunOutcome } from "#lib/test";
import { useTestStore } from "#stores/useTestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

interface SuiteRun {
	outcome: TestRunOutcome;
	at: number;
}

type TestAdapterLike = ReturnType<typeof useTestStore.getState>["adapter"];

/** runOne runs a suite file and reports the outcome. Module-level because
 * React Compiler cannot handle try/catch. */
async function runOne(
	adapter: TestAdapterLike,
	ref: TestFileRef,
	onResult: (path: string, run: SuiteRun) => void,
	onError: (message: string) => void,
): Promise<void> {
	try {
		const outcome = await adapter.run({ path: ref.path });
		onResult(ref.path, { outcome, at: Date.now() });
	} catch (err) {
		onError(err instanceof Error ? err.message : String(err));
	}
}

/** runAllSequentially runs every suite one after another (chained promises,
 * not a loop await). Module-level to keep async work out of the component. */
async function runAllSequentially(
	adapter: TestAdapterLike,
	refs: TestFileRef[],
	onResult: (path: string, run: SuiteRun) => void,
	onError: (message: string) => void,
): Promise<void> {
	await refs.reduce<Promise<void>>(
		(chain, ref) => chain.then(() => runOne(adapter, ref, onResult, onError)),
		Promise.resolve(),
	);
}

/** downloadOutcome saves a run outcome as JSON — module-level DOM work. */
function downloadOutcome(name: string, run: SuiteRun): void {
	const blob = new Blob([JSON.stringify(run.outcome, null, 2)], {
		type: "application/json",
	});
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = `${name.replace(/\.[^.]+$/, "")}-run.json`;
	a.click();
	URL.revokeObjectURL(url);
}

function relativeAge(at: number): string {
	const s = Math.max(1, Math.round((Date.now() - at) / 1000));
	if (s < 60) return `${s}s ago`;
	const m = Math.round(s / 60);
	if (m < 60) return `${m}m ago`;
	return `${Math.round(m / 60)}h ago`;
}

/** SuiteResultList renders per-test pass/fail lines for the last run. */
function SuiteResultList({ outcome }: { outcome: TestRunOutcome }) {
	if (!outcome.results) return null;
	return (
		<ul className="flex flex-col gap-0.5">
			{outcome.results.flatMap((r) =>
				r.results.map((a, j) => (
					<li
						// Results are a positional report of the last run; the run is
						// immutable once recorded, so index + test name is stable here.
						// react-doctor-disable-next-line react-doctor/no-array-index-as-key
						key={`${r.name}-${j}`}
						className={cn(
							"flex items-center gap-2 rounded border border-border/50 px-2 py-1 text-xs",
							a.passed ? "text-status-ok" : "text-status-error",
						)}
					>
						<span aria-hidden>{a.passed ? "✓" : "✗"}</span>
						<span className="min-w-0 flex-1 truncate">
							<span className="font-medium text-foreground">{r.name}</span>
							<span className="text-muted-foreground">
								{a.message ? ` — ${a.message}` : ` — ${a.assertion.kind}`}
							</span>
						</span>
					</li>
				)),
			)}
		</ul>
	);
}

/** TestsView is the G-17.4.11 tests home: suites sidebar with run-all beside
 * a per-suite runner (run, outcome summary, results, JSON export, open in
 * editor). JUnit export awaits a report-format seam. */
export function TestsView() {
	const tests = useTestStore((s) => s.tests);
	const refreshList = useTestStore((s) => s.refreshList);
	const adapter = useTestStore((s) => s.adapter);
	const openPath = useTestStore((s) => s.openPath);
	const newTab = useTestStore((s) => s.newTab);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);

	const [selectedPath, setSelectedPath] = useState<string | null>(null);
	const [runs, setRuns] = useState<Record<string, SuiteRun>>({});
	const [runningPath, setRunningPath] = useState<string | null>(null);
	const [runAll, setRunAll] = useState(false);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		void refreshList();
	}, [refreshList]);

	const selected = tests.find((t) => t.path === selectedPath) ?? null;
	const lastRun = selectedPath ? runs[selectedPath] : undefined;

	const runSuite = (ref: TestFileRef): void => {
		setRunningPath(ref.path);
		setError(null);
		void runOne(adapter, ref, (path, run) => {
			setRuns((prev) => ({ ...prev, [path]: run }));
		}, setError).finally(() => {
			setRunningPath(null);
		});
	};

	const runEverySuite = (): void => {
		setRunAll(true);
		setError(null);
		void runAllSequentially(adapter, tests, (path, run) => {
			setRuns((prev) => ({ ...prev, [path]: run }));
		}, setError).finally(() => {
			setRunAll(false);
		});
	};

	const openInEditor = (ref: TestFileRef): void => {
		const id = `test-${ref.path}`;
		void openPath(id, ref.path);
		openTab({ id, title: ref.name || ref.path, kind: "test", filePath: ref.path });
		setActiveView("requests");
	};

	const newTest = (): void => {
		const id = `test-new-${Date.now()}`;
		newTab(id);
		openTab({ id, title: "untitled.reqly-test", kind: "test" });
		setActiveView("requests");
	};

	return (
		<div className="flex h-full min-h-0 gap-4 p-4" aria-label="Tests">
			<aside className="flex w-64 shrink-0 flex-col gap-2 rounded-xl border border-border bg-card p-3">
				<div className="flex items-center justify-between">
					<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
						Test suites
					</p>
					<Button size="icon" variant="ghost" aria-label="New test file" onClick={newTest}>
						<FilePlus className="size-3.5" />
					</Button>
				</div>
				{tests.length === 0 ? (
					<p className="text-[11px] text-muted-foreground">
						No *.reqly-test files found in the workspace.
					</p>
				) : (
					<ul className="flex flex-col gap-1">
						{tests.map((t) => {
							const run = runs[t.path];
							return (
								<li key={t.path}>
									<button
										type="button"
										onClick={() => setSelectedPath(t.path)}
										className={cn(
											"w-full rounded-lg border px-2.5 py-2 text-left transition-colors",
											selectedPath === t.path
												? "border-primary/50 bg-primary/5"
												: "border-transparent hover:bg-accent",
										)}
									>
										<span className="block truncate text-xs font-medium text-foreground">
											{t.name || t.path}
										</span>
										<span className="block truncate font-mono text-[10px] text-muted-foreground">
											{t.path}
										</span>
										{run ? (
											<span
												className={cn(
													"mt-0.5 block font-data text-[10px]",
													run.outcome.passed ? "text-status-ok" : "text-status-error",
												)}
											>
												{run.outcome.passCount}/{run.outcome.total} passed
											</span>
										) : null}
									</button>
								</li>
							);
						})}
					</ul>
				)}
				<Button
					variant="outline"
					size="sm"
					className="mt-auto"
					disabled={tests.length === 0 || runAll}
					onClick={() => void runEverySuite()}
				>
					{runAll ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
					Run all suites
				</Button>
			</aside>

			<section className="flex min-w-0 flex-1 flex-col gap-3 overflow-y-auto">
				{error ? <p className="text-xs text-status-error">{error}</p> : null}
				{selected ? (
					<>
						<div className="flex flex-wrap items-center gap-2">
							<h2 className="text-sm font-semibold text-foreground">
								{selected.name || selected.path}
							</h2>
							{lastRun ? (
								<span
									className={cn(
										"rounded-full border px-2 py-0.5 font-data text-[10px]",
										lastRun.outcome.passed
											? "border-status-ok/40 text-status-ok"
											: "border-status-error/40 text-status-error",
									)}
								>
									{lastRun.outcome.passCount}/{lastRun.outcome.total} passed ·{" "}
									{(lastRun.outcome.durationMs / 1000).toFixed(1)}s ·{" "}
									{relativeAge(lastRun.at)}
								</span>
							) : null}
							<div className="ml-auto flex items-center gap-2">
								<Button
									variant="destructive"
									size="sm"
									disabled={runningPath != null}
									onClick={() => void runSuite(selected)}
								>
									{runningPath === selected.path ? (
										<Spinner data-icon="inline-start" />
									) : (
										<Play data-icon="inline-start" />
									)}
									Run suite
								</Button>
								{lastRun ? (
									<Button
										variant="outline"
										size="sm"
										onClick={() => downloadOutcome(selected.name || "suite", lastRun)}
									>
										<Download data-icon="inline-start" />
										JSON
									</Button>
								) : null}
								<Button variant="outline" size="sm" onClick={() => openInEditor(selected)}>
									<Pencil data-icon="inline-start" />
									Edit
								</Button>
							</div>
						</div>

						<p className="font-mono text-[11px] text-muted-foreground">{selected.path}</p>

						{lastRun ? (
							<SuiteResultList outcome={lastRun.outcome} />
						) : (
							<div className="flex flex-1 items-center justify-center rounded-xl border border-border bg-card">
								<p className="text-xs text-muted-foreground">
									Run the suite to see assertion results here.
								</p>
							</div>
						)}
					</>
				) : (
					<div className="flex flex-1 items-center justify-center rounded-xl border border-border bg-card">
						<p className="text-xs text-muted-foreground">
							Pick a suite to run it, or create a new *.reqly-test file.
						</p>
					</div>
				)}
			</section>
		</div>
	);
}
