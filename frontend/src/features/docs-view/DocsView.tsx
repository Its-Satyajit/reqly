import { useEffect, useState } from "react";
import { Check, Copy, FileText } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { copyText } from "#lib/response";
import { handleTabArrowKeys, tabClass } from "#lib/ui";
import { useDocsStore } from "#stores/useDocsStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

/** DocsView generates Markdown API documentation from the open workspace
 * (G-15): pick collections, render into .reqly/docs/, preview each file and
 * copy it — identical fidelity to `reqly docs generate`. */
export function DocsView() {
	const collections = useWorkspaceStore((s) => s.workspaceTree?.collections ?? []);
	const selected = useDocsStore((s) => s.selected);
	const outName = useDocsStore((s) => s.outName);
	const busy = useDocsStore((s) => s.busy);
	const error = useDocsStore((s) => s.error);
	const result = useDocsStore((s) => s.result);
	const activeFile = useDocsStore((s) => s.activeFile);
	const toggleCollection = useDocsStore((s) => s.toggleCollection);
	const selectAll = useDocsStore((s) => s.selectAll);
	const setOutName = useDocsStore((s) => s.setOutName);
	const setActiveFile = useDocsStore((s) => s.setActiveFile);
	const generate = useDocsStore((s) => s.generate);
	const [copied, setCopied] = useState(false);
	// Set membership keeps the per-collection checkbox render O(1); the
	// React Compiler handles any caching, so no manual memoization here.
	const selectedSet = new Set(selected);

	// Reset the copied affordance shortly after a successful copy.
	useEffect(() => {
		if (!copied) return;
		const t = setTimeout(() => setCopied(false), 1500);
		return () => clearTimeout(t);
	}, [copied]);

	const active = result?.files.find((f) => f.name === activeFile) ?? null;

	return (
		<div className="flex h-full min-h-0 flex-col gap-2 p-3">
			<div className="flex items-center gap-2">
				<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
					Docs
				</p>
			</div>

			<div className="flex flex-col gap-1">
				<div className="flex items-center justify-between">
					<p className="text-xs text-muted-foreground">
						Collections {selected.length === 0 && "(all)"}
					</p>
					{selected.length > 0 && (
						<button
							type="button"
							onClick={selectAll}
							className="text-xs text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
						>
							Reset to all
						</button>
					)}
				</div>
				{collections.length > 0 ? (
					<ul className="flex flex-wrap gap-x-3 gap-y-1">
						{collections.map((c) => (
							<li key={c.name}>
								<label className="flex items-center gap-1 text-xs text-foreground">
									<input
										type="checkbox"
										checked={selectedSet.has(c.name)}
										onChange={() => toggleCollection(c.name)}
										aria-label={`Include collection ${c.name}`}
										className="size-3.5 accent-(--primary)"
									/>
									{c.name}
								</label>
							</li>
						))}
					</ul>
				) : (
					<p className="py-8 text-center text-sm text-muted-foreground">No collections yet.</p>
				)}
			</div>

			<div className="flex items-center gap-1.5">
				<Input
					value={outName}
					onChange={(e) => setOutName(e.target.value)}
					placeholder="Output folder name (default: docs-<timestamp>)"
					spellCheck={false}
					aria-label="Output folder name"
					className="w-72 font-mono text-xs"
				/>
				<Button size="sm" onClick={() => void generate()} disabled={busy}>
					{busy ? <Spinner data-icon="inline-start" /> : <FileText data-icon="inline-start" />}
					Generate
				</Button>
			</div>

			{error && (
				<Alert variant="destructive">
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			)}

			{result && (
				<>
					<p className="font-data text-xs text-muted-foreground" title={result.path}>
						Saved to <span className="font-mono">{result.path}</span> ·{" "}
						{result.requestCount} requests · {result.files.length} files
					</p>
					<div
						className="flex shrink-0 items-center gap-1"
						role="tablist"
						aria-label="Generated doc files"
						onKeyDown={(e) => handleTabArrowKeys(e)}
					>
						{result.files.map((f) => (
							<button
								key={f.name}
								type="button"
								role="tab"
								aria-selected={activeFile === f.name}
								tabIndex={activeFile === f.name ? 0 : -1}
								onClick={() => setActiveFile(f.name)}
								className={tabClass(activeFile === f.name)}
							>
								{f.name}
							</button>
						))}
					</div>
					<div className="flex min-h-0 flex-1 flex-col rounded-md border border-border">
						{active ? (
							<>
								<div className="flex shrink-0 items-center justify-between border-b border-border px-2 py-1">
									<span className="text-xs text-muted-foreground">
										Markdown preview — {active.name}
									</span>
									<Button
										size="sm"
										variant="outline"
										onClick={() =>
											void copyText(active.content).then((ok) => setCopied(ok))
										}
									>
										{copied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
										{copied ? "Copied" : "Copy"}
									</Button>
								</div>
								<CodeMirrorEditor
									value={active.content}
									language="markdown"
									readOnly
									className="min-h-0 flex-1 overflow-hidden"
								/>
							</>
						) : (
							<p className="p-2 text-xs text-muted-foreground">No file selected.</p>
						)}
					</div>
				</>
			)}

			{!result && !error && (
				<p className="text-xs text-muted-foreground">
					Generate Markdown documentation for the workspace — one page per collection with methods,
					URLs, headers, bodies, and ready-to-run cURL snippets.
				</p>
			)}
		</div>
	);
}
