import { useState } from "react";
import { Sparkles, FileJson, FlaskConical, Stethoscope, BookText } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getAIBridge } from "#lib/ai";

type Tab = "explain" | "tests" | "docs" | "diagnose";

export function AiView() {
	const [responseJson, setResponseJson] = useState('{"status":200,"body":{"id":1,"name":"alice"}}');
	const [errMsg, setErrMsg] = useState("");
	const [active, setActive] = useState<Tab>("explain");
	const [output, setOutput] = useState<string>("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const run = async (tab: Tab) => {
		setActive(tab);
		setBusy(true);
		setError(null);
		setOutput("");
		try {
			let res = "";
			if (tab === "explain") res = await getAIBridge().explain(responseJson);
			else if (tab === "tests") res = await getAIBridge().generateTests(responseJson);
			else if (tab === "docs") res = await getAIBridge().generateDocs("", responseJson);
			else if (tab === "diagnose") res = await getAIBridge().diagnose(responseJson, errMsg);
			setOutput(res);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock for vite dev
				if (tab === "explain") setOutput("response 200 OK in 42ms (123B) proto HTTP/1.1 (mock)");
				else if (tab === "tests") setOutput(`reqly.test("Status code is 200", function() { return reqly.response.status === 200; });\n// mock`);
				else if (tab === "docs") setOutput("# `GET` https://api.example.com/users\n\n## Response (`200 OK` - 42ms)\n\n```json\n{\n  \"id\": 1\n}\n```\n// mock");
				else setOutput("### 🟢 Success (200 OK)\n\nRequest completed normally in 42ms. (mock)");
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="AI assistant">
			<PageHeader icon={Sparkles} title="AI Assistant" description="Local heuristics — explain, test, docs, diagnose — zero telemetry, no cloud." />

			<div className="grid gap-4 lg:grid-cols-[0.95fr_1.25fr]">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<FileJson className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Response JSON</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">drop → explain</Badge>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<Textarea
							value={responseJson}
							onChange={(e) => setResponseJson(e.target.value)}
							rows={10}
							spellCheck={false}
							className="font-mono text-xs"
							placeholder='{"status":200,"body":{"id":1}}'
						/>
						<input
							value={errMsg}
							onChange={(e) => setErrMsg(e.target.value)}
							placeholder="error message for Diagnose (optional)"
							className="h-7 rounded-md border border-input bg-transparent px-2 text-xs font-mono"
						/>
						<div className="flex flex-wrap gap-1.5">
							<Button size="sm" variant={active === "explain" ? "default" : "outline"} onClick={() => void run("explain")} disabled={busy} className="gap-1.5">
								<Sparkles className="size-3.5" /> Explain
							</Button>
							<Button size="sm" variant={active === "tests" ? "default" : "outline"} onClick={() => void run("tests")} disabled={busy} className="gap-1.5">
								<FlaskConical className="size-3.5" /> Tests
							</Button>
							<Button size="sm" variant={active === "docs" ? "default" : "outline"} onClick={() => void run("docs")} disabled={busy} className="gap-1.5">
								<BookText className="size-3.5" /> Docs
							</Button>
							<Button size="sm" variant={active === "diagnose" ? "default" : "outline"} onClick={() => void run("diagnose")} disabled={busy} className="gap-1.5">
								<Stethoscope className="size-3.5" /> Diagnose
							</Button>
						</div>
						{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<h3 className="text-sm font-semibold capitalize">{active}</h3>
						<span className="font-mono text-[11px] text-muted-foreground">· local heuristics</span>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<pre className="max-h-[420px] overflow-auto whitespace-pre-wrap break-all rounded border bg-muted/40 p-3 font-mono text-xs leading-relaxed">{output || "Output appears here — explain shows latency/status, tests synthesize Goja assertions, docs emits Markdown."}</pre>
						<p className="font-mono text-[10px] text-muted-foreground">AI is local, zero network — uses `internal/ai` templates, not an LLM.</p>
					</div>
				</div>
			</div>
		</section>
	);
}
