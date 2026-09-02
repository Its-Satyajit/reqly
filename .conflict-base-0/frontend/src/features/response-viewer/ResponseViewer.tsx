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
	{ id: "tests", label: "Tests" },
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
		<div className="flex h-full min-h-0 flex-col bg-card">
			{/* Unified header — status + view rail, no double bar */}
			<div className="flex shrink-0 flex-col border-b border-border bg-card">
				<div className="flex h-9 items-center justify-between gap-2 px-3">
					<div className="flex min-w-0 items-center gap-2">
						<span className="hidden font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground sm:inline">
							Response
						</span>
						{cancelled ? (
							<span className="rounded-md bg-status-warn/10 px-1.5 py-0.5 font-mono text-xs text-status-warn">Cancelled</span>
						) : response ? (
							<StatusPill status={response.statusCode} />
						) : (
							<span className="font-mono text-xs text-muted-foreground">No response</span>
						)}
						{response && (
							<span className="hidden items-center gap-1.5 font-mono text-xs tabular-nums text-muted-foreground sm:inline-flex">
								<span className="text-foreground/80">{response.statusText}</span>
								<span className="text-border">·</span>
								<span>{response.durationMs}ms</span>
								<span className="text-border">·</span>
								<span>{formatBytes(response.size)}</span>
								{ct && (
									<>
										<span className="text-border">·</span>
										<span className="max-w-28 truncate">{ct.split(";")[0]}</span>
									</>
								)}
							</span>
						)}
					</div>

					{response && (
						<div className="flex shrink-0 items-center gap-1">
							<Button
								size="xs"
								variant="ghost"
								className="h-7 px-2 font-mono text-[11px] text-muted-foreground hover:text-foreground"
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
								className="h-7 px-2 font-mono text-[11px] text-muted-foreground hover:text-foreground"
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
					)}
				</div>

				{response && (
					<div
						className="flex items-center justify-between gap-2 border-t border-border/50 bg-muted/20 px-2 py-1"
						role="tablist"
						aria-label="Response views"
						onKeyDown={(e) => handleTabArrowKeys(e)}
					>
						<div className="flex items-center gap-0.5 overflow-x-auto">
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
										"shrink-0 rounded-md px-2.5 py-1 text-xs font-medium transition-colors",
										view === v.id
											? "bg-background text-foreground shadow-sm"
											: "text-muted-foreground hover:bg-background/60 hover:text-foreground",
										v.id === "table" && !tabular ? "cursor-not-allowed opacity-40" : "",
									)}
								>
									{v.label}
								</button>
							))}
						</div>

						<div className="hidden shrink-0 items-center gap-1.5 sm:flex">
							{parsed !== null && (
								<input
									value={jsonPath}
									onChange={(e) => setJsonPath(e.target.value)}
									placeholder="$.data[*]"
									aria-label="JSONPath query"
									spellCheck={false}
									className="h-7 w-32 rounded-md border border-input bg-background px-2 font-mono text-xs placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none lg:w-40"
								/>
							)}
							<input
								value={query}
								onChange={(e) => setQuery(e.target.value)}
								placeholder="Filter…"
								aria-label="Search response"
								className="h-7 w-28 rounded-md border border-input bg-background px-2 font-mono text-xs placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none lg:w-32"
							/>
							{searchResult && searchResult.count > 0 ? (
								<span className="font-mono text-[11px] font-medium tabular-nums text-primary">
									{searchResult.count}
								</span>
							) : null}
						</div>
					</div>
				)}
			</div>

			{/* mobile filter row — hidden on desktop where it lives in header */}
			{response && (
				<div className="flex items-center gap-1.5 border-b border-border/50 bg-muted/10 px-2 py-1.5 sm:hidden">
					{parsed !== null && (
						<input
							value={jsonPath}
							onChange={(e) => setJsonPath(e.target.value)}
							placeholder="JSONPath $.data"
							aria-label="JSONPath query"
							spellCheck={false}
							className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 font-mono text-xs focus:border-ring focus:outline-none"
						/>
					)}
					<input
						value={query}
						onChange={(e) => setQuery(e.target.value)}
						placeholder="Filter…"
						aria-label="Search response"
						className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 font-mono text-xs focus:border-ring focus:outline-none"
					/>
				</div>
			)}

			{response && binaryType !== "none" ? (
				<div className="mx-2 mt-2 shrink-0 rounded-md border border-border bg-muted/15 px-2.5 py-2">
					{binaryType === "image" ? (
						<div className="flex flex-col gap-2">
							<p className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Image preview</p>
							<img src={imageDataUrl} alt="response" className="max-h-64 self-start rounded-md border border-border" />
						</div>
					) : binaryType === "pdf" ? (
						<p className="font-mono text-xs text-muted-foreground">PDF response — use Download.</p>
					) : (
						<div className="flex flex-col gap-1.5">
							<p className="font-mono text-xs text-muted-foreground">
								Binary · {formatBytes(response.size)} — first 4 KB
							</p>
							<pre className="max-h-36 overflow-auto whitespace-pre rounded-md border border-border/50 bg-background p-2 font-mono text-[11px] leading-snug text-muted-foreground">
								{hexPreview}
							</pre>
						</div>
					)}
				</div>
			) : null}

			{parsed === null && response && jsonPath.trim() ? (
				<p className="shrink-0 border-b border-border/50 bg-status-warn/5 px-3 py-1.5 text-xs leading-snug text-muted-foreground">
					This response is not JSON — JSONPath needs a JSON body.
				</p>
			) : null}

			<div className="min-h-0 flex-1 p-2">
				{jsonPathResult &&
				!jsonPathResult.error &&
				jsonPathResult.matches.length > 0 ? (
					<div className="flex h-full flex-col gap-1.5 overflow-y-auto rounded-md border border-border bg-background p-2">
						{jsonPathResult.matches.map((m) => (
							<JsonPathMatchRow key={m.path} match={m} />
						))}
					</div>
				) : jsonPathResult?.error ? (
					<div className="flex h-full items-start rounded-md border border-destructive/30 bg-destructive/5 p-3">
						<p className="font-mono text-xs text-destructive">{jsonPathResult.error}</p>
					</div>
				) : jsonPathResult ? (
					<div className="flex h-full items-center justify-center rounded-md border border-dashed border-border bg-muted/10 p-6">
						<p className="text-xs text-muted-foreground">No matches for this path.</p>
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
						<p className="shrink-0 px-2.5 pt-2 font-mono text-xs text-muted-foreground">
							Not JSON — showing raw body.
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
							<div className="flex h-full items-center justify-center rounded-md border border-dashed border-border bg-muted/10 p-6">
								<p className="text-xs text-muted-foreground">Not tabular — need JSON array or CSV.</p>
							</div>
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
													className="sticky top-0 z-10 border-b border-border bg-muted/40 px-2.5 py-1.5 text-left font-mono text-[11px] font-semibold uppercase tracking-wide text-muted-foreground"
												>
													{c}
												</th>
											))}
										</tr>
									</thead>
									<tbody>
										{tableData.rows.map((row, i) => (
											<tr key={`row-${row[0] ?? ''}-${i}`} className="hover:bg-muted/30">
												{row.map((cell, j) => (
													<td
														key={`cell-${tableData.columns[j] ?? j}-${j}`}
														className="break-all border-b border-border/50 px-2.5 py-1.5 font-mono text-xs"
													>
														{cell}
													</td>
												))}
											</tr>
										))}
									</tbody>
								</table>
								{tableData.rows.length >= 1000 ? <p className="border-t border-border bg-muted/20 p-2 font-mono text-xs text-muted-foreground">Showing first 1000 rows.</p> : null}
							</div>
						)}
					</div>
				) : view === "cookies" ? (
					<div className="h-full overflow-y-auto rounded-md border border-border bg-background">
						{filteredCookies.length === 0 ? (
							<div className="flex h-full flex-col items-center justify-center gap-1.5 p-6 text-center">
								<p className="text-sm font-medium">
									{cookies.length === 0
										? "No cookies set by this response."
										: "No cookies match your search."}
								</p>
								{cookies.length === 0 ? (
									<p className="max-w-sm font-mono text-xs leading-relaxed text-muted-foreground">
										Servers set cookies via <code>Set-Cookie</code>. Send a request to an endpoint that sets one.
									</p>
								) : null}
							</div>
						) : (
							<table className="w-full text-left text-xs">
								<tbody>
									{filteredCookies.map((c) => (
										<tr
											key={`${c.name}-${c.value}-${c.domain ?? ""}`}
											className="border-b border-border/50 last:border-0 hover:bg-muted/20"
										>
											<td className="px-2.5 py-1.5 font-mono font-medium">{c.name}</td>
											<td className="max-w-[20ch] break-all px-2.5 py-1.5 font-mono text-muted-foreground">
												{c.value}
											</td>
											<td className="px-2.5 py-1.5 text-muted-foreground">{c.domain ?? "—"}</td>
											<td className="px-2.5 py-1.5 font-mono text-muted-foreground">
												{c.path ?? "/"}
											</td>
											<td className="px-2.5 py-1.5 text-muted-foreground">
												{cookieExpiry(c) ?? "Session"}
											</td>
											<td className="px-2.5 py-1.5 whitespace-nowrap text-muted-foreground">
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
					<div className="h-full overflow-y-auto rounded-md border border-border bg-background">
						{filteredHeaders.length === 0 ? (
							<div className="flex h-full items-center justify-center p-6">
								<p className="font-mono text-xs text-muted-foreground">
									{query ? "No headers match your search." : "No response headers."}
								</p>
							</div>
						) : (
							<table className="w-full text-left text-xs">
								<tbody>
									{filteredHeaders.map((h) => (
										<tr
											key={`${h.key}-${h.value}`}
											className="border-b border-border/50 last:border-0 hover:bg-muted/20"
										>
											<td className="w-40 shrink-0 px-2.5 py-1.5 font-mono text-muted-foreground">
												{h.key}
											</td>
											<td className="break-all px-2.5 py-1.5 font-mono text-foreground">
												{h.value}
											</td>
										</tr>
									))}
								</tbody>
							</table>
						)}
					</div>
				) : view === "tests" ? (
					<div className="flex h-full items-center justify-center rounded-md border border-dashed border-border bg-muted/10 p-6">
						<p className="font-mono text-xs text-muted-foreground">No test results — run scripts to see output.</p>
					</div>
				) : view === "timeline" ? (
					<div className="flex h-full flex-col gap-2 overflow-y-auto">
						{response ? (
							<div className="rounded-md border border-border bg-background p-3">
								<p className="font-mono text-xs text-muted-foreground">Timeline — {response.durationMs}ms · {formatBytes(response.size)} · {response.statusCode} {response.statusText}</p>
								{response.timings ? (
									<div className="mt-3 flex flex-col gap-1.5">
										{[
											{ label: "DNS", value: response.timings.dns, color: "bg-status-info" },
											{ label: "Connect", value: response.timings.connect, color: "bg-status-info" },
											{ label: "TLS", value: response.timings.tls, color: "bg-status-warn" },
											{ label: "Request", value: response.timings.request, color: "bg-status-ok" },
											{ label: "Server", value: response.timings.server, color: "bg-status-error" },
											{ label: "Response", value: response.timings.response, color: "bg-primary" },
											{ label: "Transfer", value: response.timings.transfer, color: "bg-muted-foreground" },
										].map((p) => (
											<div key={p.label} className="flex items-center gap-2 font-mono text-xs">
												<span className="w-16 text-muted-foreground">{p.label}</span>
												<div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
													<div className={cn("h-1.5 rounded-full", p.color)} style={{ width: `${Math.min(100, (p.value / Math.max(1, response.durationMs)) * 100)}%` }} />
												</div>
												<span className="w-12 text-right tabular-nums text-muted-foreground">{p.value}ms</span>
											</div>
										))}
									</div>
								) : (
									<p className="mt-2 font-mono text-xs text-muted-foreground/70">Detailed waterfall coming soon.</p>
								)}
							</div>
						) : (
							<div className="flex h-full items-center justify-center rounded-md border border-dashed border-border bg-muted/10 p-6">
								<p className="font-mono text-xs text-muted-foreground">No timeline — send a request.</p>
							</div>
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
				<p className="shrink-0 border-t border-border/50 bg-muted/10 px-3 py-1.5 font-mono text-xs text-muted-foreground">
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
		<div className="flex flex-col gap-1 rounded-md border border-border bg-muted/10 px-2.5 py-2">
			<p className="font-mono text-xs text-muted-foreground">{match.path}</p>
			<pre className="overflow-x-auto whitespace-pre-wrap font-mono text-xs leading-snug text-foreground">
				{text}
			</pre>
		</div>
	);
}
