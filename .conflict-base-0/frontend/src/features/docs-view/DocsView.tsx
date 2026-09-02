import { useEffect, useState } from "react";
import { Check, Copy, FileText } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { copyText } from "#lib/response";
import { handleTabArrowKeys } from "#lib/ui";
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
		<div className="flex h-full min-h-0 flex-col overflow-y-auto bg-background">
			<PageHeader
				icon={FileText}
				title="API Docs"
				description="Generate Markdown documentation from workspace collections with cURL snippets"
			/>
			<div className="flex flex-col gap-3 p-4">
				<div className="flex flex-col gap-1.5 rounded border border-border/80 bg-card/30 p-3">
					<div className="flex items-center justify-between">
						<span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
							Collections {selected.length === 0 ? "(All selected)" : `(${selected.length} selected)`}
						</span>
						{selected.length > 0 && (
							<button
								type="button"
								onClick={selectAll}
								className="font-mono text-[11px] text-primary hover:underline"
							>
								Select all
							</button>
						)}
					</div>
					{collections.length > 0 ? (
						<ul className="flex flex-wrap gap-2 pt-1">
							{collections.map((c) => (
								<li key={c.name}>
									<label className="flex cursor-pointer items-center gap-1.5 rounded border border-border bg-background px-2 py-1 text-xs text-foreground transition-colors hover:bg-muted select-none">
										<input
											type="checkbox"
											checked={selectedSet.has(c.name)}
											onChange={() => toggleCollection(c.name)}
											aria-label={`Include collection ${c.name}`}
											className="size-3.5 accent-(--primary)"
										/>
										<span className="font-mono text-[11px] font-medium">{c.name}</span>
									</label>
								</li>
							))}
						</ul>
					) : (
						<p className="font-mono text-xs text-muted-foreground">No collections found in workspace.</p>
					)}
				</div>

				<div className="flex items-center gap-2">
					<Input
						value={outName}
						onChange={(e) => setOutName(e.target.value)}
						placeholder="Output folder name (default: docs-<timestamp>)"
						spellCheck={false}
						aria-label="Output folder name"
						className="w-80 font-mono text-xs"
					/>
					<Button size="sm" onClick={() => void generate()} disabled={busy} className="h-8 gap-1.5 px-3 font-mono text-xs font-semibold">
						{busy ? <Spinner data-icon="inline-start" /> : <FileText className="size-3.5" aria-hidden />}
						<span>Generate Docs</span>
					</Button>
				</div>

				{error && (
					<Alert variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				)}

				{result && (
					<div className="flex flex-col gap-2 pt-2">
						<div className="flex items-center gap-2 font-mono text-[11px] text-muted-foreground">
							<span>Saved to <strong className="text-foreground">{result.path}</strong></span>
							<span>|</span>
							<span className="tabular-nums">{result.requestCount} requests</span>
							<span>|</span>
							<span className="tabular-nums">{result.files.length} files generated</span>
						</div>
						<div
							className="flex shrink-0 items-center gap-1 border-b border-border/70 pb-1"
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
									className={`rounded px-2.5 py-1 font-mono text-[11px] transition-colors ${
										activeFile === f.name
											? "bg-primary/10 font-bold text-primary"
											: "text-muted-foreground hover:bg-muted hover:text-foreground"
									}`}
								>
									{f.name}
								</button>
							))}
						</div>
						<div className="flex min-h-[350px] flex-col rounded border border-border bg-card/20">
							{active ? (
								<>
									<div className="flex shrink-0 items-center justify-between border-b border-border px-3 py-1.5 bg-background/50">
										<span className="font-mono text-xs text-muted-foreground">
											{active.name}
										</span>
										<Button
											size="xs"
											variant="ghost"
											onClick={() =>
												void copyText(active.content).then((ok) => setCopied(ok))
											}
											className="gap-1 font-mono text-[11px]"
										>
											{copied ? <Check className="size-3" /> : <Copy className="size-3" />}
											{copied ? "Copied" : "Copy Markdown"}
										</Button>
									</div>
									<CodeMirrorEditor
										value={active.content}
										language="markdown"
										readOnly
										className="min-h-[300px] flex-1 overflow-hidden"
									/>
								</>
							) : (
								<p className="p-4 font-mono text-xs text-muted-foreground">No file selected.</p>
							)}
						</div>
					</div>
				)}

				{!result && !error && (
					<p className="font-mono text-xs text-muted-foreground">
						Select collections and click Generate to produce GitHub-flavored Markdown docs with cURL snippets.
					</p>
				)}
			</div>
		</div>
	);
}
