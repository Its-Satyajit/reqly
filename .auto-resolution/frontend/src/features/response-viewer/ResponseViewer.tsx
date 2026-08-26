import { useEffect, useState } from "react";
import { Play, Zap } from "lucide-react";
import { JsonTree } from "../../components/JsonTree";
import { StatusPill } from "../../components/status";
import { Button } from "../../components/ui/button";
import { CodeMirrorEditor } from "../../editors";
import { formatBytes, tabClass } from "../../lib/ui";
import { Tabs, TabsList, TabsTrigger } from "#components/ui/tabs";
import { Skeleton } from "#components/ui/skeleton";
import { Kbd } from "#components/ui/kbd";
import { ScrollArea } from "#components/ui/scroll-area";
import { type JSONPathMatch, type JSONPathResult, queryJSONPath } from "../../lib/jsonpath";
import { isRecord, type JsonValue } from "../../lib/typeGuards";
import { methodTintClass } from "../../lib/status";
import { sentRows } from "../../lib/request";
import { cn } from "../../lib/utils";
import {
	binaryPreviewType,
	bytesToBase64,
	contentType,
	cookieExpiry,
	copyText,
	headerRows,
	hexDump,
	isTabular,
	parseSetCookies,
	parseTable,
	prettyBody,
	searchBody,
	suggestedFilename,
} from "../../lib/response";
import { notifyError } from "../../lib/notify";
import { useRequestStore } from "../../stores/useRequestStore";
import { effectiveUrlFor, useWorkspaceStore } from "../../stores/useWorkspaceStore";
import { useHistoryStore } from "../../stores/useHistoryStore";
import type { HistoryEntry } from "../../lib/history";

type View = "raw" | "pretty" | "headers" | "tree" | "cookies" | "table";

const views: { id: View; label: string }[] = [
	{ id: "raw", label: "Raw" },
	{ id: "pretty", label: "Pretty" },
	{ id: "headers", label: "Headers" },
	{ id: "tree", label: "Tree" },
	{ id: "cookies", label: "Cookies" },
	{ id: "table", label: "Table" },
];

type ResponseData = NonNullable<
	ReturnType<typeof useRequestStore.getState>["responses"][string]
>["response"];

/** Parses a JSON response body. Module-level because React Compiler cannot
 * handle try/catch. */
function parseJsonBody(body: string): JsonValue | null {
	try {
		// SAFETY: JSON response body parsed at I/O boundary; validated as JsonValue via isRecord/Array checks in viewers
		return JSON.parse(body) as JsonValue;
	} catch {
		return null;
	}
}

function cookieText(c: {
	name: string;
	value: string;
	domain?: string | null;
	path?: string | null;
	secure?: boolean;
	httpOnly?: boolean;
}): string {
	return `${c.name}=${c.value}; ${[
		c.domain ? `Domain=${c.domain}` : "",
		c.path ? `Path=${c.path}` : "",
		c.secure ? "Secure" : "",
		c.httpOnly ? "HttpOnly" : "",
	]
		.filter(Boolean)
		.join("; ")}`;
}

function ResponseMeta({
	response,
	cancelled,
	ct,
}: {
	response: ResponseData;
	cancelled: boolean;
	ct: string;
}) {
	if (cancelled) {
		return (
			<p className="font-data text-xs text-muted-foreground">Request cancelled</p>
		);
	}
	if (!response) return null;
	return (
		<p className="flex min-w-0 items-center gap-2 font-data text-xs tabular-nums text-muted-foreground">
			<StatusPill status={response.statusCode} />
			<span className="truncate">
				{response.proto ? `${response.proto} · ` : ""}
				{response.statusText}
			</span>
			<span aria-hidden>·</span>
			<span>{response.durationMs}ms</span>
			{(response.attempts ?? 1) > 1 ? (
				<>
					<span aria-hidden>·</span>
					<span title="Sends including automatic retries">
						{response.attempts} attempts
					</span>
				</>
			) : null}
			<span aria-hidden>·</span>
			<span>{formatBytes(response.size)}</span>
			{ct ? (
				<>
					<span aria-hidden>·</span>
					<span className="truncate">{ct.split(";")[0]}</span>
				</>
			) : null}
		</p>
	);
}

function BinaryPreviewBlock({
	response,
	binaryType,
	imageDataUrl,
	hexPreview,
}: {
	response: ResponseData;
	binaryType: string;
	imageDataUrl: string;
	hexPreview: string;
}) {
	if (!response || binaryType === "none") return null;
	return (
		<div className="mx-2 mb-1 shrink-0 rounded-md border border-border bg-muted/30 px-2 py-1 text-xs">
			{binaryType === "image" ? (
				<div className="flex flex-col gap-1">
					<p className="text-muted-foreground">Image preview</p>
					<img src={imageDataUrl} alt="response" className="max-h-64 rounded" />
				</div>
			) : binaryType === "pdf" ? (
				<p className="text-muted-foreground">PDF response — use Download.</p>
			) : (
				<div className="flex flex-col gap-1">
					<p className="text-muted-foreground">
						Binary ({formatBytes(response.size)}) — first 4KB + Download.
					</p>
					<pre className="max-h-40 overflow-auto whitespace-pre rounded bg-background p-2 font-mono text-[11px] leading-snug text-muted-foreground">
						{hexPreview}
					</pre>
				</div>
			)}
		</div>
	);
}

function ViewerToolbar({
	view,
	setView,
	tabular,
	query,
	setQuery,
	searchCount,
}: {
	view: View;
	setView: (v: View) => void;
	tabular: boolean;
	query: string;
	setQuery: (q: string) => void;
	searchCount: number | null;
}) {
	return (
		<div className="flex shrink-0 items-center gap-1 px-2 pb-1">
			<Tabs
				value={view}
				onValueChange={(v) => {
					// SAFETY: view ids come from the local `views` array
					setView(v as View)
				}}
			>
				<TabsList variant="line" aria-label="Response views">
					{views.map((v) => (
						<TabsTrigger
							key={v.id}
							value={v.id}
							disabled={v.id === "table" && !tabular}
							title={v.id === "table" && !tabular ? "Not tabular — need JSON array or CSV" : undefined}
							className={`${tabClass(view === v.id)} ${v.id === "table" && !tabular ? "opacity-50" : ""}`}
						>
							{v.label}
						</TabsTrigger>
					))}
				</TabsList>
			</Tabs>
			<input
				value={query}
				onChange={(e) => setQuery(e.target.value)}
				placeholder="Search response…"
				aria-label="Search response"
				className="ml-auto w-44 rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground"
			/>
			{searchCount !== null && searchCount > 0 ? (
				<span className="shrink-0 font-data text-xs tabular-nums text-muted-foreground">
					{searchCount} matches
				</span>
			) : null}
		</div>
	);
}

function TableView({ tableData }: { tableData: NonNullable<ReturnType<typeof parseTable>> }) {
	return (
		<div className="flex h-full min-h-0 flex-col">
			<div className="min-h-0 flex-1 overflow-auto rounded-md border border-border bg-background">
				<table className="w-full border-separate border-spacing-0 text-left text-xs">
					<thead>
						<tr>
							{tableData.columns.map((c) => (
								<th
									key={c}
									className="sticky top-0 z-10 border-b border-border bg-background px-2 py-1 font-medium"
								>
									{c}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{tableData.rows.map((row) => (
							<tr key={row.join("¦")}>
								{row.map((cell, j) => (
									<td
										key={`${cell}:${j}`}
										className="break-all border-b border-border/50 px-2 py-1 font-mono"
									>
										{cell}
									</td>
								))}
							</tr>
						))}
					</tbody>
				</table>
				{tableData.rows.length >= 1000 ? <p className="p-2 text-xs text-muted-foreground">Showing first 1000 rows.</p> : null}
			</div>
		</div>
	);
}

function CookiesView({
	cookies,
	filteredCookies,
}: {
	cookies: ReturnType<typeof parseSetCookies>;
	filteredCookies: ReturnType<typeof parseSetCookies>;
}) {
	return (
		<div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
			{filteredCookies.length === 0 ? (
				<div className="flex h-full flex-col items-start justify-center gap-2 px-4">
					<p className="text-sm font-medium text-foreground">
						{cookies.length === 0
							? "No cookies set by this response."
							: "No cookies match your search."}
					</p>
					{cookies.length === 0 ? (
						<p className="max-w-sm text-xs text-muted-foreground">
							Servers set cookies via{" "}
							<code className="font-mono">Set-Cookie</code> response
							headers. Send a request to an endpoint that sets a cookie to
							see it here — persistence is a separate roadmap item.
						</p>
					) : null}
				</div>
			) : (
				<table className="w-full text-left text-xs">
					<tbody>
						{filteredCookies.map((c) => (
							<tr
								key={`${c.name}-${c.value}-${c.domain ?? ""}-${c.path ?? ""}`}
								className="border-b border-border/50 last:border-0"
							>
								<td className="py-1 pr-3 align-top font-mono text-foreground">
									{c.name}
								</td>
								<td className="py-1 pr-3 font-mono text-muted-foreground break-all">
									{c.value}
								</td>
								<td className="py-1 pr-3 align-top text-muted-foreground">
									{c.domain ?? "—"}
								</td>
								<td className="py-1 pr-3 align-top font-mono text-muted-foreground">
									{c.path ?? "/"}
								</td>
								<td className="py-1 pr-3 align-top text-muted-foreground">
									{cookieExpiry(c) ?? "Session"}
								</td>
								<td className="py-1 align-top whitespace-nowrap text-muted-foreground">
									{[
										c.secure ? "Secure" : "",
										c.httpOnly ? "HttpOnly" : "",
										c.sameSite ? `SameSite=${c.sameSite}` : "",
									]
										.filter(Boolean)
										.join(" · ") || "—"}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</div>
	);
}

function HeadersView({
	headers,
	query,
}: {
	headers: ReturnType<typeof headerRows>;
	query: string;
}) {
	return (
		<ScrollArea className="h-full rounded-md border border-border bg-background p-2">
			{headers.length === 0 ? (
				<p className="text-xs text-muted-foreground">
					{query
						? "No headers match your search."
						: "No response headers."}
				</p>
			) : (
				<table className="w-full text-left text-xs">
					<tbody>
						{headers.map((h) => (
							<tr
								key={`${h.key}-${h.value}`}
								className="border-b border-border/50 last:border-0"
							>
								<td className="py-1 pr-3 align-top font-mono text-muted-foreground">
									{h.key}
								</td>
								<td className="py-1 font-mono text-foreground break-all">
									{h.value}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</ScrollArea>
	);
}

/** recentChipLabel renders a history entry as a short chip label: the last
 * URL path segment, falling back to the request path's file name. */
function recentChipLabel(entry: HistoryEntry): string {
	const withoutQuery = entry.url.split("?")[0];
	const segments = withoutQuery.split("/").filter(Boolean);
	const last = segments[segments.length - 1];
	if (last) return last;
	const file = entry.requestPath.split("/").pop() ?? entry.requestPath;
	return file.replace(/\.(json|yaml|yml)$/i, "");
}

/** ReadyHero is the "Ready to send" empty state: a bolt icon, hint, Send
 * CTA, and recent-request chips pulled from history. */
function ReadyHero({
	onSend,
	recent,
}: {
	onSend: () => void;
	recent: HistoryEntry[];
}) {
	const seen = new Set<string>();
	const chips = recent
		.filter((e) => {
			const key = `${e.method} ${e.url}`;
			if (seen.has(key)) return false;
			seen.add(key);
			return true;
		})
		.slice(0, 4);
	return (
		<div className="flex h-full items-center justify-center p-4">
			<div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-xl border border-border bg-muted/20 px-6 py-8">
				<span
					aria-hidden
					className="flex size-12 items-center justify-center rounded-xl border border-primary/25 bg-primary/10 text-primary"
				>
					<Zap className="size-5" />
				</span>
				<div className="flex flex-col items-center gap-1.5 text-center">
					<p className="text-base font-semibold text-foreground">Ready to send</p>
					<p className="max-w-xs text-xs text-muted-foreground">
						Press{" "}
						<kbd className="rounded border border-border bg-background px-1 font-data text-[10px]">
							⌘↩
						</Kbd>{" "}
						or hit Send — pre-request scripts run first, then the response
						lands here.
					</p>
				</div>
				<Button size="sm" variant="destructive" onClick={onSend}>
					<Play className="size-3 fill-current" aria-hidden />
					Send request
				</Button>
				{chips.length > 0 ? (
					<div className="flex w-full flex-col items-center gap-2 border-t border-border/50 pt-3">
						<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground/70">
							Recent
						</p>
						<div className="flex flex-wrap justify-center gap-1.5">
							{chips.map((e) => (
								<span
									key={`${e.method}-${e.url}`}
									className="flex items-center gap-1.5 rounded-full border border-border bg-background px-2 py-0.5 font-data text-[10px] text-muted-foreground"
								>
									<span
										className={cn(
											"font-semibold uppercase",
											methodTintClass(e.method),
										)}
									>
										{e.method}
									</span>
									<span className="max-w-32 truncate">{recentChipLabel(e)}</span>
								</span>
							))}
						</div>
					</div>
				) : null}
			</div>
		</div>
	);
}

/** ResponseActions is the Copy / Copy headers / Download / Format row plus
 * the JSONPath input. */
function ResponseActions({
	view,
	setView,
	headersText,
	cookiesText,
	bodyView,
	raw,
	ct,
	filename,
	parsed,
	jsonPath,
	setJsonPath,
}: {
	view: View;
	setView: (v: View) => void;
	headersText: string;
	cookiesText: string;
	bodyView: string;
	raw: string;
	ct: string;
	filename: string;
	parsed: JsonValue | null;
	jsonPath: string;
	setJsonPath: (q: string) => void;
}) {
	const [copied, setCopied] = useState(false);
	const flashCopied = () => {
		setCopied(true);
		setTimeout(() => setCopied(false), 1500);
	};
	const copy = (text: string) => {
		void copyText(text).then((ok) => {
			if (!ok) {
				notifyError("Copy failed", "Clipboard access was denied.");
				return;
			}
			flashCopied();
		});
	};
	return (
		<div className="flex shrink-0 items-center gap-1 border-t border-border/50 px-2 py-1">
			<Button
				size="xs"
				variant="ghost"
				onClick={() =>
					copy(
						view === "headers"
							? headersText
							: view === "cookies"
								? cookiesText
								: bodyView,
					)
				}
			>
				{copied ? "Copied" : "Copy"}
			</Button>
			<Button size="xs" variant="ghost" onClick={() => copy(headersText)}>
				Copy headers
			</Button>
			<Button
				size="xs"
				variant="ghost"
				onClick={() => {
					const blob = new Blob([raw], {
						type: ct || "application/octet-stream",
					});
					const url = URL.createObjectURL(blob);
					const a = document.createElement("a");
					a.href = url;
					a.download = filename;
					a.click();
					URL.revokeObjectURL(url);
				}}
			>
				Download
			</Button>
			<Button size="xs" variant="ghost" onClick={() => setView("pretty")}>
				Format
			</Button>
			<span className="ml-auto flex items-center gap-1 text-xs text-muted-foreground">
				<span>JSONPath</span>
				<input
					value={jsonPath}
					onChange={(e) => setJsonPath(e.target.value)}
					placeholder="$.users[*].name"
					aria-label="JSONPath query"
					disabled={parsed === null}
					spellCheck={false}
					className="w-48 rounded-md border border-input bg-background px-2 py-1 font-mono text-xs text-foreground placeholder:text-muted-foreground disabled:opacity-50"
				/>
			</span>
		</div>
	);
}

export function ResponseViewer() {
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const draft = useRequestStore((s) =>
		activeTabId ? s.drafts[activeTabId] : undefined,
	);
	const meta = useRequestStore((s) =>
		activeTabId ? s.meta[activeTabId] : undefined,
	);
	const send = useRequestStore((s) => s.send);
	const tabState = useRequestStore((s) =>
		activeTabId ? s.responses[activeTabId] : undefined,
	);
	const historyPool = useHistoryStore((s) => s.pool);
	const poolLoaded = useHistoryStore((s) => s.poolLoaded);
	const loadPool = useHistoryStore((s) => s.loadPool);
	const response = tabState?.response ?? null;
	const loading = tabState?.loading ?? false;
	const error = tabState?.error ?? null;
	const cancelled = tabState?.cancelled ?? false;
	const [view, setView] = useState<View>("pretty");
	const [query, setQuery] = useState("");
	const [jsonPath, setJsonPath] = useState("");

	useEffect(() => {
		if (!poolLoaded) void loadPool();
	}, [poolLoaded, loadPool]);

	const resolvedUrl = draft
		? effectiveUrlFor(draft.url, meta?.baseUrl ?? "")
		: "";

	const handleSend = () => {
		if (!activeTabId || !draft) return;
		void send(activeTabId, {
			method: draft.method,
			url: draft.url,
			params: sentRows(draft.params),
			headers: sentRows(draft.headers).map(({ key, value }) => ({ key, value })),
			bodyType: draft.bodyType,
			body: draft.body,
			form: sentRows(draft.form),
			graphqlQuery: draft.graphqlQuery,
			graphqlVariables: draft.graphqlVariables,
			env: meta?.env,
			requestPath: meta?.requestPath,
			auth: draft.auth,
		});
	};

	const ct = response ? contentType(response.headers) : "";
	const pretty = response ? prettyBody(response.body, ct) : "";
	const raw = response?.body ?? "";
	const parsed = response ? parseJsonBody(response.body) : null;
	const bodyView = view === "pretty" ? pretty : view === "raw" ? raw : "";
	const treeFallback = view === "tree" && parsed === null && response != null;
	const headers = response ? headerRows(response.headers) : [];

	const filename = response ? suggestedFilename(response.headers, ct) : "";
	const headersText = headers.map((h) => `${h.key}: ${h.value}`).join("\n");
	const cookies = response ? parseSetCookies(response.headers) : [];
	const cookiesText = cookies.map(cookieText).join("\n");
	const tabular = response ? isTabular(response.body, ct) : false;
	const tableData =
		response && view === "table" ? parseTable(response.body, ct) : null;
	const binaryType = binaryPreviewType(ct);
	const jsonPathResult: JSONPathResult | null =
		parsed && jsonPath.trim() ? queryJSONPath(parsed, jsonPath) : null;
	const hexPreview =
		response && binaryType === "hex" ? hexDump(raw) : "";
	const imageDataUrl =
		response && binaryType === "image" && raw
			? `data:${ct.split(";")[0] || "application/octet-stream"};base64,${bytesToBase64(raw)}`
			: "";

	const body = loading
		? "// Sending request…"
		: cancelled
			? "// Request cancelled"
			: error
				? `// Error: ${error}`
				: response
					? bodyView
					: "// Send a request to see the response";

	const searchResult = searchBody(bodyView, query);
	const filteredHeaders =
		view === "headers" && query.trim()
			? headers.filter(
					(h) =>
						h.key.toLowerCase().includes(query.toLowerCase()) ||
						h.value.toLowerCase().includes(query.toLowerCase()),
				)
			: headers;

	const filteredCookies =
		view === "cookies" && query.trim()
			? cookies.filter((c) =>
					`${c.name} ${c.value} ${c.domain ?? ""} ${c.path ?? ""}`
						.toLowerCase()
						.includes(query.toLowerCase()),
				)
			: cookies;

	return (
		<div className="flex h-full min-h-0 flex-col">
			<div className="flex items-center justify-between gap-2 px-2 pb-1 pt-2">
				<div className="flex min-w-0 items-baseline gap-2">
					<p className="shrink-0 text-xs font-medium uppercase tracking-wide text-muted-foreground">
						Response
					</p>
					{resolvedUrl ? (
						<p className="min-w-0 truncate font-data text-xs text-muted-foreground/70">
							{resolvedUrl}
						</p>
					) : null}
				</div>
				<ResponseMeta response={response} cancelled={cancelled} ct={ct} />
			</div>

			<BinaryPreviewBlock
				response={response}
				binaryType={binaryType}
				imageDataUrl={imageDataUrl}
				hexPreview={hexPreview}
			/>

			{response ? (
				<ViewerToolbar
					view={view}
					setView={setView}
					tabular={tabular}
					query={query}
					setQuery={setQuery}
					searchCount={searchResult?.count ?? null}
				/>
			) : null}

			{response ? (
				<ResponseActions
					view={view}
					setView={setView}
					headersText={headersText}
					cookiesText={cookiesText}
					bodyView={bodyView}
					raw={raw}
					ct={ct}
					filename={filename}
					parsed={parsed}
					jsonPath={jsonPath}
					setJsonPath={setJsonPath}
				/>
			) : null}

			{parsed === null && response && jsonPath.trim() ? (
				<p className="shrink-0 border-t border-border/50 px-2 py-1 text-xs text-muted-foreground">
					This response is not JSON — JSONPath queries need a JSON body.
				</p>
			) : null}

			<div className="min-h-0 flex-1 p-2 pt-0">
				{jsonPathResult &&
				!jsonPathResult.error &&
				jsonPathResult.matches.length > 0 ? (
					<div className="flex h-full flex-col gap-1 overflow-y-auto rounded-md border border-border bg-background p-2">
						{jsonPathResult.matches.map((m) => (
							<JsonPathMatchRow key={m.path} match={m} />
						))}
					</div>
				) : jsonPathResult?.error ? (
					<div className="flex h-full items-start rounded-md border border-border bg-background p-2">
						<p className="text-xs text-destructive">{jsonPathResult.error}</p>
					</div>
				) : jsonPathResult ? (
					<div className="flex h-full items-start rounded-md border border-border bg-background p-2">
						<p className="text-xs text-muted-foreground">
							No matches for this path.
						</p>
					</div>
				) : loading ? (
					<div
						className="flex h-full flex-col gap-2 rounded-md border border-border bg-background p-3"
						role="status"
						aria-label="Loading response"
					>
						<span className="sr-only">Sending request…</span>
						{[100, 92, 96, 78, 88].map((w, i) => (
							// SAFETY: static skeleton widths — positional identity is the point
							<Skeleton
								key={w}
								className="h-3"
								style={{ width: `${w}%`, animationDelay: `${i * 120}ms` }}
							/>
						))}
					</div>
				) : error ? (
					<CodeMirrorEditor
						value={body}
						language="text"
						readOnly
						className="h-full overflow-hidden rounded-md border border-border"
					/>
				) : view === "tree" && parsed !== null ? (
					<ScrollArea className="h-full rounded-md border border-border bg-background p-2">
						<JsonTree data={parsed} filter={query} />
					</ScrollArea>
				) : view === "tree" && treeFallback ? (
					<div className="flex h-full min-h-0 flex-col rounded-md border border-border">
						<p className="shrink-0 px-2 pt-2 text-xs text-muted-foreground">
							This response is not JSON — showing the raw body.
						</p>
						<CodeMirrorEditor
							value={raw}
							language="text"
							readOnly
							className="min-h-0 flex-1 overflow-hidden"
						/>
					</div>
				) : view === "table" ? (
					!tabular ? (
						<p className="text-xs text-muted-foreground">Not tabular — need JSON array or CSV.</p>
					) : !tableData || tableData.columns.length === 0 ? (
						<p className="text-xs text-muted-foreground">No tabular data.</p>
					) : (
						<TableView tableData={tableData} />
					)
				) : view === "cookies" ? (
					<CookiesView cookies={cookies} filteredCookies={filteredCookies} />
				) : view === "headers" ? (
					<HeadersView headers={filteredHeaders} query={query} />
				) : response ? (
					<CodeMirrorEditor
						value={body}
						language="json"
						readOnly
						className="h-full overflow-hidden rounded-md border border-border"
					/>
				) : (
					<ReadyHero onSend={handleSend} recent={historyPool} />
				)}
			</div>

			{view !== "tree" &&
			view !== "headers" &&
			view !== "cookies" &&
			response &&
			query &&
			searchResult &&
			searchResult.count === 0 ? (
				<p className="shrink-0 px-2 pb-1 text-xs text-muted-foreground">
					No matches in the response body.
				</p>
			) : null}
		</div>
	);
}

function JsonPathMatchRow({ match }: { match: JSONPathMatch }) {
	const text =
		match.value === null
			? "null"
			: isRecord(match.value) || Array.isArray(match.value)
				? JSON.stringify(match.value, null, 2)
				: String(match.value);
	return (
		<div className="flex flex-col gap-0.5 rounded-md border border-border/50 bg-background px-2 py-1">
			<p className="font-mono text-xs text-muted-foreground">{match.path}</p>
			<pre className="overflow-x-auto whitespace-pre-wrap font-mono text-xs text-foreground">
				{text}
			</pre>
		</div>
	);
}
