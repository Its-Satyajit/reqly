import { useMemo, useState } from "react";
import { JsonTree } from "../../components/JsonTree";
import { StatusPill } from "../../components/status";
import { Button } from "../../components/ui/button";
import { CodeMirrorEditor } from "../../editors";
import { formatBytes, handleTabArrowKeys } from "../../lib/ui";
import { cn } from "#lib/utils";
import { type JSONPathMatch, queryJSONPath } from "../../lib/jsonpath";
import { isRecord, type JsonValue } from "../../lib/typeGuards";
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
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";

type View = "raw" | "pretty" | "headers" | "tree" | "cookies" | "table" | "tests" | "timeline";

const views: { id: View; label: string }[] = [
	{ id: "pretty", label: "Body" },
	{ id: "raw", label: "Raw" },
	{ id: "headers", label: "Headers" },
	{ id: "cookies", label: "Cookies" },
	{ id: "tree", label: "Tree" },
	{ id: "table", label: "Table" },
	{ id: "tests", label: "Test Results" },
	{ id: "timeline", label: "Timeline" },
];

export function ResponseViewer() {
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const tabState = useRequestStore((s) =>
		activeTabId ? s.responses[activeTabId] : undefined,
	);
	const response = tabState?.response ?? null;
	const loading = tabState?.loading ?? false;
	const error = tabState?.error ?? null;
	const cancelled = tabState?.cancelled ?? false;
	const [view, setView] = useState<View>("pretty");
	const [query, setQuery] = useState("");
	const [jsonPath, setJsonPath] = useState("");
	const [copied, setCopied] = useState(false);

	const ct = response ? contentType(response.headers) : "";
	const pretty = useMemo(
		() => (response ? prettyBody(response.body, ct) : ""),
		[response, ct],
	);
	const raw = response?.body ?? "";
	const parsed = useMemo(() => {
		if (!response) return null;
		try {
			// SAFETY: JSON response body parsed at I/O boundary; validated as JsonValue via isRecord/Array checks in viewers
			return JSON.parse(response.body) as JsonValue;
		} catch {
			return null;
		}
	}, [response]);
	const bodyView = view === "pretty" ? pretty : view === "raw" ? raw : "";
	const treeFallback = view === "tree" && parsed === null && response != null;
	const headers = response ? headerRows(response.headers) : [];

	const filename = response ? suggestedFilename(response.headers, ct) : "";
	const headersText = headers.map((h) => `${h.key}: ${h.value}`).join("\n");
	const cookies = useMemo(
		() => (response ? parseSetCookies(response.headers) : []),
		[response],
	);
	const cookiesText = cookies
		.map(
			(c) =>
				`${c.name}=${c.value}; ${[
					c.domain ? `Domain=${c.domain}` : "",
					c.path ? `Path=${c.path}` : "",
					c.secure ? "Secure" : "",
					c.httpOnly ? "HttpOnly" : "",
				]
					.filter(Boolean)
					.join("; ")}`,
		)
		.join("\n");
	const tabular = useMemo(() => (response ? isTabular(response.body, ct) : false), [response, ct]);
	const tableData = useMemo(() => (response && view === "table" ? parseTable(response.body, ct) : null), [response, view, ct]);
	const binaryType = useMemo(() => binaryPreviewType(ct), [ct]);
	const jsonPathResult = useMemo(() => {
		if (!parsed || !jsonPath.trim()) return null;
		return queryJSONPath(parsed, jsonPath);
	}, [parsed, jsonPath]);
	const hexPreview = useMemo(
		() => (response && binaryType === "hex" ? hexDump(raw) : ""),
		[response, binaryType, raw],
	);
	const imageDataUrl = useMemo(
		() =>
			response && binaryType === "image" && raw
				? `data:${ct.split(";")[0] || "application/octet-stream"};base64,${bytesToBase64(raw)}`
				: "",
		[response, binaryType, raw, ct],
	);

	const body = loading
		? "// Sending request…"
		: cancelled
			? "// Request cancelled"
			: error
				? `// Error: ${error}`
				: response
					? bodyView
					: "// Send a request to see the response";

	const searchResult = useMemo(() => searchBody(bodyView, query), [bodyView, query]);
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
		<div className="flex h-full min-h-0 flex-col bg-background">
			<div className="flex h-10 shrink-0 items-center justify-between border-b border-border bg-card/20 px-3 select-none">
				<div className="flex items-center gap-2">
					<span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
						Response
					</span>
					{cancelled ? (
						<span className="font-mono text-xs text-status-warn">Request cancelled</span>
					) : response ? (
						<StatusPill status={response.statusCode} />
					) : null}
				</div>

				{response && (
					<div className="flex items-center gap-2.5 font-mono text-[11px] tabular-nums text-muted-foreground">
						<span className="text-foreground/80">{response.statusText}</span>
						<span className="text-border">|</span>
						<span>{response.durationMs}ms</span>
						<span className="text-border">|</span>
						<span>{formatBytes(response.size)}</span>
						{ct && (
							<>
								<span className="text-border">|</span>
								<span className="truncate max-w-32">{ct.split(";")[0]}</span>
							</>
						)}
					</div>
				)}
			</div>

			{response && binaryType !== "none" ? (
				<div className="mx-2 my-1 shrink-0 rounded border border-border bg-muted/20 px-2 py-1 text-xs">
					{binaryType === "image" ? (
						<div className="flex flex-col gap-1">
							<p className="text-muted-foreground font-mono text-[11px]">Image preview</p>
							<img src={imageDataUrl} alt="response" className="max-h-64 rounded border border-border/50" />
						</div>
					) : binaryType === "pdf" ? (
						<p className="text-muted-foreground font-mono text-[11px]">PDF response — use Download.</p>
					) : (
						<div className="flex flex-col gap-1">
							<p className="text-muted-foreground font-mono text-[11px]">
								Binary ({formatBytes(response.size)}) — first 4KB + Download.
							</p>
							<pre className="max-h-40 overflow-auto whitespace-pre rounded bg-background p-2 font-mono text-[11px] leading-snug text-muted-foreground border border-border/50">
								{hexPreview}
							</pre>
						</div>
					)}
				</div>
			) : null}

			{response ? (
				<div
					className="flex shrink-0 items-center justify-between border-b border-border/70 bg-muted/10 px-2.5 py-1 select-none"
					role="tablist"
					aria-label="Response views"
					onKeyDown={(e) => handleTabArrowKeys(e)}
				>
					<div className="flex items-center gap-1">
						{views.map((v) => (
							<button
								key={v.id}
								type="button"
								role="tab"
								aria-selected={view === v.id}
								tabIndex={view === v.id ? 0 : -1}
								onClick={() => setView(v.id)}
								disabled={v.id === "table" && !tabular}
								title={v.id === "table" && !tabular ? "Not tabular — need JSON array or CSV" : undefined}
								className={cn(
									"rounded px-2 py-0.5 font-mono text-[11px] transition-colors",
									view === v.id
										? "bg-background font-semibold text-primary shadow-xs"
										: "text-muted-foreground hover:bg-muted hover:text-foreground",
									v.id === "table" && !tabular ? "opacity-40 cursor-not-allowed" : "",
								)}
							>
								{v.label}
							</button>
						))}
					</div>

					<div className="flex items-center gap-1.5">
						{parsed !== null && (
							<input
								value={jsonPath}
								onChange={(e) => setJsonPath(e.target.value)}
								placeholder="JSONPath (e.g. $.data[*])"
								aria-label="JSONPath query"
								spellCheck={false}
								className="h-6 w-32 lg:w-44 rounded border border-input bg-background px-2 font-mono text-[11px] text-foreground placeholder:text-muted-foreground/60 focus:border-ring focus:outline-none"
							/>
						)}
						<input
							value={query}
							onChange={(e) => setQuery(e.target.value)}
							placeholder="Filter body…"
							aria-label="Search response"
							className="h-6 w-24 lg:w-32 rounded border border-input bg-background px-2 font-mono text-[11px] text-foreground placeholder:text-muted-foreground/60 focus:border-ring focus:outline-none"
						/>
						{searchResult && searchResult.count > 0 ? (
							<span className="font-mono text-[10px] tabular-nums text-primary font-medium">
								{searchResult.count}
							</span>
						) : null}

						<div className="h-3.5 w-px bg-border/80 mx-0.5" />

						<Button
							size="xs"
							variant="ghost"
							className="h-6 px-1.5 text-[11px] font-mono text-muted-foreground hover:text-foreground"
							onClick={() => {
								const text =
									view === "headers"
										? headersText
										: view === "cookies"
											? cookiesText
											: bodyView;
								void copyText(text).then((ok) => {
									if (!ok) {
										notifyError("Copy failed", "Clipboard access was denied.");
										return;
									}
									setCopied(true);
									setTimeout(() => setCopied(false), 1500);
								});
							}}
						>
							{copied ? "Copied" : "Copy"}
						</Button>
						<Button
							size="xs"
							variant="ghost"
							className="h-6 px-1.5 text-[11px] font-mono text-muted-foreground hover:text-foreground"
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
					</div>
				</div>
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
							<div
								key={i}
								className="h-3 animate-pulse rounded bg-muted"
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
					<div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
						<JsonTree data={parsed} filter={query} />
					</div>
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
					<div className="flex h-full min-h-0 flex-col">
						{!tabular ? (
							<p className="text-xs text-muted-foreground">Not tabular — need JSON array or CSV.</p>
						) : !tableData || tableData.columns.length === 0 ? (
							<p className="text-xs text-muted-foreground">No tabular data.</p>
						) : (
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
										{tableData.rows.map((row, i) => (
											<tr key={`row-${row[0] ?? ''}-${i}`}>
												{row.map((cell, j) => (
													<td
														key={`cell-${tableData.columns[j] ?? j}-${j}`}
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
						)}
					</div>
				) : view === "cookies" ? (
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
											key={`${c.name}-${c.value}-${c.domain ?? ""}`}
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
				) : view === "headers" ? (
					<div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
						{filteredHeaders.length === 0 ? (
							<p className="text-xs text-muted-foreground">
								{query
									? "No headers match your search."
									: "No response headers."}
							</p>
						) : (
							<table className="w-full text-left text-xs">
								<tbody>
									{filteredHeaders.map((h) => (
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
					</div>
				) : view === "tests" ? (
					<div className="flex h-full items-center justify-center p-4">
						<p className="text-xs text-muted-foreground">No test results — run pre-request/tests scripts to see output.</p>
					</div>
				) : view === "timeline" ? (
					<div className="flex h-full flex-col gap-2 p-2">
						{response ? (
							<div className="rounded-md border border-border bg-background p-2 text-xs">
								<p className="text-muted-foreground">Timeline — request took {response.durationMs}ms · {formatBytes(response.size)} · {response.statusCode} {response.statusText}</p>
								{response.timings ? (
									<div className="mt-2 flex flex-col gap-1">
										{[
											{ label: "DNS", value: response.timings.dns, color: "bg-status-info" },
											{ label: "Connect", value: response.timings.connect, color: "bg-status-info" },
											{ label: "TLS", value: response.timings.tls, color: "bg-status-warn" },
											{ label: "Request", value: response.timings.request, color: "bg-status-ok" },
											{ label: "Server", value: response.timings.server, color: "bg-status-error" },
											{ label: "Response", value: response.timings.response, color: "bg-primary" },
											{ label: "Transfer", value: response.timings.transfer, color: "bg-muted-foreground" },
										].map((p) => (
											<div key={p.label} className="flex items-center gap-2">
												<span className="w-16 text-muted-foreground">{p.label}</span>
												<div className="h-2 flex-1 rounded bg-muted">
													<div className={`h-2 rounded ${p.color}`} style={{ width: `${Math.min(100, (p.value / Math.max(1, response.durationMs)) * 100)}%` }} />
												</div>
												<span className="w-12 text-right tabular-nums">{p.value}ms</span>
											</div>
										))}
									</div>
								) : (
									<p className="mt-1 text-muted-foreground/70">Detailed waterfall coming soon.</p>
								)}
							</div>
						) : (
							<p className="text-xs text-muted-foreground">No timeline — send a request.</p>
						)}
					</div>
				) : response ? (
					<CodeMirrorEditor
						value={body}
						language="json"
						readOnly
						className="h-full overflow-hidden rounded-md border border-border"
					/>
				) : (
					<CodeMirrorEditor
						value={body}
						language="text"
						readOnly
						className="h-full overflow-hidden rounded-md border border-border"
					/>
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
	const text = useMemo(
		() =>
			match.value === null
				? "null"
				: isRecord(match.value) || Array.isArray(match.value)
					? JSON.stringify(match.value, null, 2)
					: String(match.value),
		[match.value],
	);
	return (
		<div className="flex flex-col gap-0.5 rounded-md border border-border/50 bg-background px-2 py-1">
			<p className="font-mono text-xs text-muted-foreground">{match.path}</p>
			<pre className="overflow-x-auto whitespace-pre-wrap font-mono text-xs text-foreground">
				{text}
			</pre>
		</div>
	);
}
