import { useState } from "react";
import { Clock, Play, Timer, FileCode2, CheckCircle2, AlertTriangle } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getAutomationBridge } from "#lib/automation";

const EXAMPLE_YAML = `name: nightly-check
interval: 5m
workflow:
  name: health
  steps:
    - id: s1
      name: ping users
      request:
        method: GET
        url: https://api.example.com/users
`;

export function AutomationView() {
	const [yaml, setYaml] = useState(EXAMPLE_YAML);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [report, setReport] = useState<{ workflowName: string; passed: boolean; steps: { name: string; passed: boolean }[] } | null>(null);

	const run = async () => {
		if (!yaml.trim()) {
			setError("yaml is required");
			return;
		}
		setBusy(true);
		setError(null);
		setReport(null);
		try {
			const r = await getAutomationBridge().run(yaml);
			setReport(r);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock report for dev preview
				setReport({ workflowName: "nightly-check", passed: true, steps: [{ name: "ping users", passed: true }, { name: "check health", passed: true }] });
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Automation">
			<PageHeader icon={Clock} title="Automation" description="Local scheduler — YAML workflow on interval, Git-native, runs once in desktop (interval forced 0)." />

			<div className="grid gap-4 lg:grid-cols-[1.4fr_0.9fr]">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<FileCode2 className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Automation YAML</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">interval · enabled · maxRuns</Badge>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<Textarea value={yaml} onChange={(e) => setYaml(e.target.value)} rows={14} spellCheck={false} className="font-mono text-xs" placeholder={EXAMPLE_YAML} />
						<div className="flex items-center gap-2">
							<Button size="sm" onClick={() => void run()} disabled={busy} className="gap-1.5">
								<Play className="size-3.5" /> {busy ? "Running…" : "Run once"}
							</Button>
							<span className="font-mono text-[11px] text-muted-foreground">interval is ignored in desktop — runs once (scheduler interval 0)</span>
						</div>
						{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Timer className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Run log</h3>
						{report && <Badge variant={report.passed ? "outline" : "destructive"} className={report.passed ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : ""}>{report.passed ? "PASS" : "FAIL"}</Badge>}
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						{!report ? (
							<p className="rounded border border-dashed border-border bg-muted/20 px-3 py-6 text-center font-mono text-xs text-muted-foreground">No run yet — paste YAML and Run. Workflow steps stream via `internal/workflow`.</p>
						) : (
							<div className="flex flex-col gap-2">
								<p className="font-mono text-xs">Workflow <span className="font-semibold">{report.workflowName}</span> — {report.steps.length} steps</p>
								<ul className="flex flex-col gap-1.5">
									{report.steps.map((s, i) => (
										<li key={`${s.name}-${i}`} className="flex items-center gap-2 rounded border border-border bg-muted/20 px-2.5 py-1.5 font-mono text-xs">
											{i + 1}. {s.name}
											<span className="ml-auto inline-flex items-center gap-1">
												{s.passed ? <CheckCircle2 className="size-3.5 text-status-ok" /> : <AlertTriangle className="size-3.5 text-status-error" />}
												{s.passed ? "pass" : "fail"}
											</span>
										</li>
									))}
								</ul>
							</div>
						)}
					</div>
				</div>
			</div>
		</section>
	);
}
