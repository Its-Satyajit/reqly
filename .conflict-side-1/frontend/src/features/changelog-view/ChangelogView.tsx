import { useState } from "react";
import { FileDiff, AlertTriangle, Sparkles, Wrench, Copy } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
import { copyText } from "#lib/response";
import { getChangelogBridge } from "#lib/changelog";
import type { Changelog } from "#lib/changelog";

function SemVerPill({ bump }: { bump: string }) {
	const cls =
		bump === "major" ? "border-status-error/20 bg-status-error/10 text-status-error" : bump === "minor" ? "border-status-redirect/20 bg-status-redirect/10 text-status-redirect" : bump === "patch" ? "border-status-warn/20 bg-status-warn/10 text-status-warn" : "border-border bg-muted text-muted-foreground";
	return <Badge variant="outline" className={`font-mono text-[11px] ${cls}`}>{bump}</Badge>;
}

export function ChangelogView() {
	const [oldPath, setOldPath] = useState("specs/v1.yaml");
	const [newPath, setNewPath] = useState("specs/v2.yaml");
	const [format, setFormat] = useState<"markdown" | "json">("markdown");
	const [failOnBreaking, setFailOnBreaking] = useState(false);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [result, setResult] = useState<{ changelog: Changelog; markdown: string; json: string } | null>(null);

	const generate = async () => {
		if (!oldPath.trim() || !newPath.trim()) {
			setError("old and new spec paths are required");
			return;
		}
		setBusy(true);
		setError(null);
		setResult(null);
		try {
			const r = await getChangelogBridge().generate(oldPath.trim(), newPath.trim(), format, failOnBreaking);
			setResult(r);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock changelog for dev
				const mock: Changelog = {
					suggested_semver: "minor",
					breaking: [],
					additions: [{ type: "create", path: "paths./users.get", summary: "Added `paths./users.get`", severity: "addition" }],
					info: [{ type: "update", path: "info.version", summary: "Modified `info.version` (1.0.0 -> 1.1.0)", severity: "info" }],
				};
				setResult({ changelog: mock, markdown: "# API Changelog\n\n**Suggested Version Bump:** `minor`\n\n### ✨ Additions\n\n- Added `paths./users.get`\n", json: JSON.stringify(mock, null, 2) });
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Changelog">
			<PageHeader icon={FileDiff} title="Changelog" description="Human-readable API diff — Old vs New spec, SemVer bump, breaking/features/fixes. Local oasdiff, no cloud." />

			<div className="grid gap-4 lg:grid-cols-[0.95fr_1.25fr]">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<FileDiff className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Specs</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">oasdiff · 0600</Badge>
					</div>
					<div className="flex flex-col gap-2 px-3 pb-3">
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Old spec</span>
							<Input value={oldPath} onChange={(e) => setOldPath(e.target.value)} placeholder="specs/v1.yaml — workspace-relative or absolute" className="font-mono text-xs" />
						</label>
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">New spec</span>
							<Input value={newPath} onChange={(e) => setNewPath(e.target.value)} placeholder="specs/v2.yaml" className="font-mono text-xs" />
						</label>

						<div className="grid grid-cols-2 gap-2">
							<label className="flex flex-col gap-1">
								<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Format</span>
								<select
									value={format}
									onChange={(e) => {
										// SAFETY: format is constrained to two values via select options
										setFormat(e.target.value as "markdown" | "json");
									}}
									className="h-7 rounded-md border border-input bg-transparent px-2 text-xs"
								>
									<option value="markdown">markdown</option>
									<option value="json">json</option>
								</select>
							</label>
							<label className="flex items-center gap-2 pt-5 font-mono text-xs">
								<input type="checkbox" checked={failOnBreaking} onChange={(e) => setFailOnBreaking(e.target.checked)} className="size-3.5 accent-[var(--primary)]" />
								fail on breaking
							</label>
						</div>

						<Button size="sm" onClick={() => void generate()} disabled={busy} className="w-fit">
							{busy ? "Generating…" : "Generate"}
						</Button>

						{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<h3 className="text-sm font-semibold">Result</h3>
						{result && <SemVerPill bump={result.changelog.suggested_semver} />}
						<span className="ml-auto flex items-center gap-1">
							<Button variant="ghost" size="xs" onClick={() => void copyText(result ? (format === "markdown" ? result.markdown : result.json) : "")} disabled={!result} className="gap-1">
								<Copy className="size-3.5" /> Copy
							</Button>
						</span>
					</div>

					<div className="flex flex-col gap-3 px-3 pb-3">
						{!result ? (
							<p className="rounded border border-dashed border-border bg-muted/20 px-3 py-8 text-center font-mono text-xs text-muted-foreground">Pick two specs and Generate — changelog is pure `internal/diffing`, no network.</p>
						) : (
							<>
								<div className="grid grid-cols-3 gap-2">
									<div className="rounded border border-status-error/20 bg-status-error/10 p-2 text-center">
										<p className="font-mono text-[11px] font-semibold text-status-error">Breaking</p>
										<p className="font-mono text-lg font-bold text-status-error">{result.changelog.breaking.length}</p>
									</div>
									<div className="rounded border border-status-redirect/20 bg-status-redirect/10 p-2 text-center">
										<p className="font-mono text-[11px] font-semibold text-status-redirect">Features</p>
										<p className="font-mono text-lg font-bold text-status-redirect">{result.changelog.additions.length}</p>
									</div>
									<div className="rounded border border-status-warn/20 bg-status-warn/10 p-2 text-center">
										<p className="font-mono text-[11px] font-semibold text-status-warn">Fixes/Info</p>
										<p className="font-mono text-lg font-bold text-status-warn">{result.changelog.info.length}</p>
									</div>
								</div>

								{result.changelog.breaking.length > 0 && (
									<div className="rounded border border-status-error/20 bg-card p-2">
										<p className="flex items-center gap-1 font-mono text-xs font-semibold text-status-error"><AlertTriangle className="size-3.5" /> Breaking</p>
										<ul className="mt-1.5 flex flex-col gap-1 font-mono text-xs">
											{result.changelog.breaking.map((i) => (
												<li key={i.path} className="rounded bg-muted/40 px-2 py-1">{i.summary}</li>
											))}
										</ul>
									</div>
								)}
								{result.changelog.additions.length > 0 && (
									<div className="rounded border border-status-redirect/20 bg-card p-2">
										<p className="flex items-center gap-1 font-mono text-xs font-semibold text-status-redirect"><Sparkles className="size-3.5" /> Additions</p>
										<ul className="mt-1.5 flex flex-col gap-1 font-mono text-xs">
											{result.changelog.additions.map((i) => (
												<li key={i.path} className="rounded bg-muted/40 px-2 py-1">{i.summary}</li>
											))}
										</ul>
									</div>
								)}
								{result.changelog.info.length > 0 && (
									<div className="rounded border border-border bg-card p-2">
										<p className="flex items-center gap-1 font-mono text-xs font-semibold"><Wrench className="size-3.5" /> Other</p>
										<ul className="mt-1.5 flex flex-col gap-1 font-mono text-xs">
											{result.changelog.info.map((i) => (
												<li key={i.path} className="rounded bg-muted/40 px-2 py-1">{i.summary}</li>
											))}
										</ul>
									</div>
								)}

								<details className="rounded border border-border/60 bg-muted/20 p-2">
									<summary className="cursor-pointer font-mono text-xs font-medium">Raw {format}</summary>
									<pre className="mt-2 max-h-48 overflow-auto whitespace-pre-wrap break-all rounded bg-background p-2 font-mono text-[11px]">{format === "markdown" ? result.markdown : result.json}</pre>
								</details>
							</>
						)}
					</div>
				</div>
			</div>
		</section>
	);
}
