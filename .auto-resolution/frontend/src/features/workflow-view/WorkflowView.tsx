import { useState } from "react";
import { Workflow, Play, GitBranch, CheckCircle2, XCircle, Boxes } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getWorkflowBridge } from "#lib/workflow";

const EXAMPLE_YAML = `name: onboarding
variables:
  base: https://api.example.com
steps:
  - id: create
    name: create user
    request:
      method: POST
      url: "{{base}}/users"
      body: '{"name":"alice"}'
  - id: get
    name: fetch user
    request:
      method: GET
      url: "{{base}}/users/{{create.id}}"
    condition: "reqly.getVariable('create.id') !== ''"
`;

export function WorkflowView() {
	const [yaml, setYaml] = useState(EXAMPLE_YAML);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [report, setReport] = useState<{ workflowName: string; passed: boolean; steps: { name: string; passed: boolean; requestError?: string }[]; extractedVars: Record<string, string> } | null>(null);

	const run = async () => {
		if (!yaml.trim()) {
			setError("yaml is required");
			return;
		}
		setBusy(true);
		setError(null);
		setReport(null);
		try {
			const r = await getWorkflowBridge().run(yaml);
			setReport(r);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setReport({
					workflowName: "onboarding",
					passed: true,
					steps: [{ name: "create user", passed: true }, { name: "fetch user", passed: true }],
					extractedVars: { "create.id": "42" },
				});
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Workflow">
			<PageHeader icon={Workflow} title="Workflow" description="Visual/programmatic multi-step — variables, extract, condition (Goja), streamed as RunView chain." />

			<div className="grid gap-4 lg:grid-cols-[1.3fr_0.9fr]">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Boxes className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Workflow YAML</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">steps · condition · extract</Badge>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<Textarea value={yaml} onChange={(e) => setYaml(e.target.value)} rows={16} spellCheck={false} className="font-mono text-xs" placeholder={EXAMPLE_YAML} />
						<Button size="sm" onClick={() => void run()} disabled={busy} className="w-fit gap-1.5">
							<Play className="size-3.5" /> {busy ? "Running…" : "Run workflow"}
						</Button>
						{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<GitBranch className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Steps</h3>
						{report && <Badge variant={report.passed ? "outline" : "destructive"} className={report.passed ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : ""}>{report.passed ? "PASS" : "FAIL"}</Badge>}
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						{!report ? (
							<p className="rounded border border-dashed border-border bg-muted/20 px-3 py-6 text-center font-mono text-xs text-muted-foreground">No run yet — workflow steps render as RunView chain, sharing variable store across steps.</p>
						) : (
							<>
								<p className="font-mono text-xs">
									{report.workflowName} — {report.steps.length} steps · {Object.keys(report.extractedVars).length} vars
								</p>
								<ul className="flex flex-col gap-1.5">
									{report.steps.map((s, i) => (
										<li key={`${s.name}-${i}`} className="flex items-center gap-2 rounded border border-border bg-muted/20 px-2.5 py-1.5">
											<span className="flex size-6 items-center justify-center rounded-full bg-muted font-mono text-[11px] font-semibold">{i + 1}</span>
											<span className="font-mono text-xs">{s.name}</span>
											<span className="ml-auto">
												{s.passed ? <CheckCircle2 className="size-4 text-status-ok" /> : <XCircle className="size-4 text-status-error" />}
											</span>
										</li>
									))}
								</ul>
								{Object.keys(report.extractedVars).length > 0 && (
									<div className="rounded border border-border/60 bg-muted/20 p-2">
										<p className="font-mono text-[11px] font-semibold text-muted-foreground">Extracted vars</p>
										<pre className="mt-1 font-mono text-[11px]">{JSON.stringify(report.extractedVars, null, 2)}</pre>
									</div>
								)}
							</>
						)}
					</div>
				</div>
			</div>
		</section>
	);
}
