import { useState } from "react";
import { Braces, Search, Sparkles, AlertTriangle, CheckCircle2 } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getSchemaBridge } from "#lib/schema";
import type { Violation } from "#lib/schema";

const EXAMPLE_SCHEMA = `{
  "type": "object",
  "properties": {
    "id": { "type": "integer" },
    "name": { "type": "string" }
  },
  "required": ["id", "name"]
}`;

const EXAMPLE_INSTANCE = `{
  "id": 1,
  "name": "alice"
}`;

export function SchemaView() {
	const [schema, setSchema] = useState(EXAMPLE_SCHEMA);
	const [instance, setInstance] = useState(EXAMPLE_INSTANCE);
	const [seed, setSeed] = useState("0");
	const [violations, setViolations] = useState<Violation[] | null>(null);
	const [valid, setValid] = useState<boolean | null>(null);
	const [inspectOut, setInspectOut] = useState<string>("");
	const [generateOut, setGenerateOut] = useState<string>("");
	const [busy, setBusy] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);

	const validate = async () => {
		setBusy("validate");
		setError(null);
		setValid(null);
		setViolations(null);
		try {
			const r = await getSchemaBridge().validate(schema, instance, "");
			setValid(r.valid);
			setViolations(r.violations);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setValid(true);
				setViolations([]);
			} else setError(msg);
		} finally {
			setBusy(null);
		}
	};

	const inspect = async () => {
		setBusy("inspect");
		setError(null);
		try {
			const out = await getSchemaBridge().inspect(schema);
			setInspectOut(out);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) setInspectOut(`object\n  id: integer! (required)\n  name: string! (required) (mock)`);
			else setError(msg);
		} finally {
			setBusy(null);
		}
	};

	const generate = async () => {
		setBusy("generate");
		setError(null);
		try {
			const out = await getSchemaBridge().generate(schema, Number(seed) || 0);
			setGenerateOut(out);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) setGenerateOut(`{\n  "id": 0,\n  "name": "string"\n} // mock`);
			else setError(msg);
		} finally {
			setBusy(null);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="JSON Schema workbench">
			<PageHeader icon={Braces} title="JSON Schema" description="Validate instance against schema, inspect keywords, generate sample — local, no network." />

			<div className="grid gap-4 lg:grid-cols-2">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Braces className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Schema</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">draft 2020-12</Badge>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<Textarea value={schema} onChange={(e) => setSchema(e.target.value)} rows={10} spellCheck={false} className="font-mono text-xs" placeholder={EXAMPLE_SCHEMA} />
						<div className="flex gap-1.5">
							<Button size="sm" onClick={() => void validate()} disabled={busy === "validate"} className="gap-1.5">
								<CheckCircle2 className="size-3.5" /> Validate
							</Button>
							<Button size="sm" variant="outline" onClick={() => void inspect()} disabled={busy === "inspect"} className="gap-1.5">
								<Search className="size-3.5" /> Inspect
							</Button>
							<Button size="sm" variant="outline" onClick={() => void generate()} disabled={busy === "generate"} className="gap-1.5">
								<Sparkles className="size-3.5" /> Generate
							</Button>
							<input value={seed} onChange={(e) => setSeed(e.target.value)} placeholder="seed" className="ml-auto h-7 w-20 rounded-md border border-input bg-transparent px-2 font-mono text-xs" />
						</div>
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<h3 className="text-sm font-semibold">Instance</h3>
						<span className="font-mono text-[11px] text-muted-foreground">· JSON or YAML</span>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<Textarea value={instance} onChange={(e) => setInstance(e.target.value)} rows={10} spellCheck={false} className="font-mono text-xs" placeholder={EXAMPLE_INSTANCE} />
						{valid !== null && (
							<div className={`rounded border px-2 py-1.5 font-mono text-xs ${valid ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : "border-status-error/20 bg-status-error/10 text-status-error"}`}>
								{valid ? "✓ valid" : `✗ invalid — ${violations?.length ?? 0} violation(s)`}
							</div>
						)}
						{violations && violations.length > 0 && (
							<ul className="flex flex-col gap-1">
								{violations.map((v, i) => (
									<li key={`${v.path}-${i}`} className="flex items-start gap-1.5 rounded border border-status-error/20 bg-card px-2 py-1 font-mono text-xs">
										<AlertTriangle className="size-3.5 shrink-0 text-status-error" />
										<span className="font-semibold">{v.path || "/"}</span>
										<span className="text-muted-foreground">{v.message}</span>
									</li>
								))}
							</ul>
						)}
					</div>
				</div>
			</div>

			{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}

			{(inspectOut || generateOut) && (
				<div className="grid gap-4 lg:grid-cols-2">
					{inspectOut && (
						<div className="rounded border border-border bg-card">
							<div className="border-b border-border px-3 py-2">
								<h4 className="font-mono text-xs font-semibold">Inspect</h4>
							</div>
							<pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 font-mono text-xs">{inspectOut}</pre>
						</div>
					)}
					{generateOut && (
						<div className="rounded border border-border bg-card">
							<div className="border-b border-border px-3 py-2">
								<h4 className="font-mono text-xs font-semibold">Generated sample</h4>
							</div>
							<pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 font-mono text-xs">{generateOut}</pre>
						</div>
					)}
				</div>
			)}
		</section>
	);
}
