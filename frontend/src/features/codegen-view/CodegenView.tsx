import { useState } from "react";
import { Check, Code, Copy, Download } from "lucide-react";
import { Button } from "#components/ui/button";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { cn } from "#lib/utils";
import { methodTintClass } from "#lib/status";
import { generateCode, type CodeLang } from "#lib/codegen";
import { copyText } from "#lib/response";
import { notifyError } from "#lib/notify";
import { useRequestStore, type TabDraft } from "#stores/useRequestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { SplitView, ViewShell } from "../../components/shell/ViewLayout";

const LANGS: { id: CodeLang; label: string }[] = [
	{ id: "curl", label: "cURL" },
	{ id: "js", label: "JavaScript" },
	{ id: "python", label: "Python" },
	{ id: "go", label: "Go" },
];

const EDITOR_LANGUAGE = {
	curl: "text",
	js: "javascript",
	python: "text",
	go: "text",
} satisfies Record<CodeLang, "javascript" | "text">;

/** codegenInput projects a tab's draft into the generator input. */
function codegenInput(
	draft: TabDraft,
	meta: ReturnType<typeof useRequestStore.getState>["meta"][string] | undefined,
	includeAuth: boolean,
) {
	let url = draft.url;
	if (meta?.baseUrl && !url.includes("://")) {
		url = `${meta.baseUrl}/${url.replace(/^\/+/, "")}`;
	}
	const headers = draft.headers.reduce<{ key: string; value: string }[]>((acc, h) => {
		if (h.enabled && h.key.trim() !== "") acc.push({ key: h.key, value: h.value });
		return acc;
	}, []);
	const query = draft.params.reduce<{ key: string; value: string }[]>((acc, p) => {
		if (p.enabled && p.key.trim() !== "") acc.push({ key: p.key, value: p.value });
		return acc;
	}, []);
	const formPairs = draft.form.reduce<string[]>((acc, f) => {
		if (f.enabled && f.key.trim() !== "") {
			acc.push(`${encodeURIComponent(f.key)}=${encodeURIComponent(f.value)}`);
		}
		return acc;
	}, []);
	const body =
		draft.bodyType === "none"
			? ""
			: draft.bodyType === "form-data" || draft.bodyType === "urlencoded"
				? formPairs.join("&")
				: draft.bodyType === "graphql"
					? draft.graphqlQuery
					: draft.body;
	return {
		method: draft.method,
		url,
		headers,
		query,
		body,
		auth: includeAuth ? draft.auth : undefined,
	};
}

/** CodegenView is the G-17.4.14 code-generation page (reference
 * 15-code-generation.png): source/language pickers beside a live output
 * pane with Copy/Download. PHP/Rust await generator seams. */
export function CodegenView() {
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const drafts = useRequestStore((s) => s.drafts);
	const meta = useRequestStore((s) => s.meta);
	const [tabId, setTabId] = useState<string | null>(activeTabId);
	const [lang, setLang] = useState<CodeLang>("curl");
	const [includeAuth, setIncludeAuth] = useState(true);
	const [copied, setCopied] = useState(false);

	const candidates = openTabs.filter((t) => drafts[t.id] != null);
	const selectedId = tabId != null && drafts[tabId] != null ? tabId : candidates[0]?.id ?? null;
	const draft = selectedId ? drafts[selectedId] : undefined;
	const selectedMeta = selectedId ? meta[selectedId] : undefined;
	const selectedTab = openTabs.find((t) => t.id === selectedId);

	const code = draft
		? generateCode(codegenInput(draft, selectedMeta, includeAuth), lang)
		: "";

	const copy = (): void => {
		void copyText(code).then((ok) => {
			if (!ok) {
				notifyError("Copy failed", "Clipboard access was denied.");
				return;
			}
			setCopied(true);
			setTimeout(() => setCopied(false), 1500);
		});
	};

	const download = (): void => {
		const ext = lang === "python" ? "py" : lang === "go" ? "go" : lang === "js" ? "js" : "sh";
		const blob = new Blob([code], { type: "text/plain" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `${(selectedTab?.title ?? "request").replace(/\.[^.]+$/, "")}.${ext}`;
		a.click();
		URL.revokeObjectURL(url);
	};

	return (
		<ViewShell label="Code generation">
			<SplitView
				asideLabel="Generator options"
				asideClassName="border-t border-border pt-3"
				aside={
						<>
					<div className="flex flex-col gap-1">
						<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
							Source
						</p>
						<select
							value={selectedId ?? ""}
							onChange={(e) => setTabId(e.target.value)}
							aria-label="Source request"
							className="h-8 rounded-md border border-border bg-transparent px-2 text-xs"
						>
							{candidates.length === 0 ? <option value="">No open requests</option> : null}
							{candidates.map((t) => (
								<option key={t.id} value={t.id}>
									{t.title}
								</option>
							))}
						</select>
					</div>
					<div className="flex flex-col gap-1">
						<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
							Language
						</p>
						<div role="radiogroup" aria-label="Language" className="flex flex-col gap-1">
							{LANGS.map((l) => (
								<button
									key={l.id}
									type="button"
									role="radio"
									aria-checked={lang === l.id}
									onClick={() => setLang(l.id)}
									className={cn(
										"flex items-center gap-2 rounded-lg border px-2.5 py-1.5 text-left text-xs transition-colors",
										lang === l.id
											? "border-primary/50 bg-primary/5 font-medium text-primary"
											: "border-border text-muted-foreground hover:bg-accent hover:text-foreground",
									)}
								>
									<Code className="size-3.5" aria-hidden />
									{l.label}
								</button>
							))}
						</div>
					</div>
					<div className="flex flex-col gap-1">
						<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
							Options
						</p>
						<label className="flex items-center gap-1.5 text-xs text-muted-foreground">
							<input
								type="checkbox"
								checked={includeAuth}
								onChange={(e) => setIncludeAuth(e.target.checked)}
								className="size-3.5 accent-(--primary)"
							/>
							include authentication
						</label>
						<p className="text-xs text-muted-foreground/70">
							Secrets are emitted as-is from the draft — never stored in generated code.
						</p>
					</div>
					</>
				}
			>
			<section className="flex min-w-0 flex-1 flex-col gap-2">
				<div className="flex flex-wrap items-center gap-2">
					{draft ? (
						<>
							<span
								className={cn(
									"shrink-0 rounded-full border border-border bg-muted/40 px-2 py-0.5 font-data text-2xs font-semibold uppercase",
									methodTintClass(draft.method),
								)}
							>
								{draft.method}
							</span>
							<span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground">
								{draft.url || "—"}
							</span>
							<span className="shrink-0 rounded-full border border-primary/25 bg-primary/10 px-2 py-0.5 font-data text-2xs text-primary">
								{LANGS.find((l) => l.id === lang)?.label}
							</span>
						</>
					) : (
						<span className="text-xs text-muted-foreground">
							Open a request tab to generate client code from it.
						</span>
					)}
					<div className="ml-auto flex shrink-0 items-center gap-2">
						<Button variant="outline" size="sm" onClick={copy} disabled={!draft}>
							{copied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
							{copied ? "Copied" : "Copy"}
						</Button>
						<Button variant="outline" size="sm" onClick={download} disabled={!draft}>
							<Download data-icon="inline-start" />
							Download
						</Button>
					</div>
				</div>
				<div className="min-h-0 flex-1 overflow-hidden rounded-xl border border-border">
					<CodeMirrorEditor
						value={code}
						language={EDITOR_LANGUAGE[lang]}
						readOnly
						className="h-full"
					/>
				</div>
			</section>
			</SplitView>
		</ViewShell>
	);
}
