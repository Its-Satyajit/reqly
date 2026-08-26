import { useState } from "react";
import { Play, RefreshCw, Settings } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Tabs, TabsList, TabsTrigger } from "#components/ui/tabs";
import { StatusPill } from "#components/status";
import { CodeMirrorEditor } from "../../editors";
import { KeyValueEditor } from "#components/KeyValueEditor";
import { JsonTree } from "#components/JsonTree";
import { cn } from "#lib/utils";
import { formatBytes, tabClass } from "#lib/ui";
import { prettyBody } from "#lib/response";
import type { KeyValueRow, RequestSender, ResponseData } from "#lib/request";
import type { JsonValue } from "#lib/typeGuards";
import {
	getGqlBridge,
	gqlTypeRef,
	type GqlField,
	type GqlSchema,
	type GqlType,
} from "#lib/graphql";
import { useRequestStore } from "#stores/useRequestStore";

type EditorTab = "query" | "variables" | "headers";

/** parseJsonBody parses a JSON response body. Module-level because React
 * Compiler cannot handle try/catch. */
function parseJsonBody(body: string): JsonValue | null {
	try {
		// SAFETY: response body parsed at the I/O boundary; validated as
		// JsonValue via isRecord/Array checks before rendering.
		return JSON.parse(body) as JsonValue;
	} catch {
		return null;
	}
}

const TABS: { id: EditorTab; label: string }[] = [
	{ id: "query", label: "Query" },
	{ id: "variables", label: "Variables" },
	{ id: "headers", label: "Headers" },
];

interface GqlOperation {
	name: string;
	type: string;
}

/** parseOperations extracts declared operations for the <auto> dropdown.
 * Module-level because React Compiler cannot handle try/catch. */
function parseOperations(query: string): GqlOperation[] {
	const ops: GqlOperation[] = [];
	const re = /^\s*(query|mutation|subscription)\s+([A-Za-z_]\w*)?/gm;
	let m: RegExpExecArray | null;
	while ((m = re.exec(query)) !== null) {
		ops.push({ type: m[1], name: m[2] ?? "(anonymous)" });
	}
	return ops;
}

/** prettifyQuery re-indents a GraphQL document by brace/paren depth.
 * Best-effort formatting for the Prettify action. */
function prettifyQuery(query: string): string {
	const lines: string[] = [];
	let depth = 0;
	for (const raw of query.split("\n")) {
		const trimmed = raw.trim();
		if (trimmed === "") continue;
		const closers = trimmed.startsWith("}") || trimmed.startsWith(")");
		if (closers) depth = Math.max(0, depth - 1);
		lines.push("  ".repeat(depth) + trimmed);
		const opens = (trimmed.match(/\{/g)?.length ?? 0) - (trimmed.match(/\}/g)?.length ?? 0);
		if (opens > 0) depth += opens;
	}
	return `${lines.join("\n")}\n`;
}

/** runQuery sends the GraphQL document through the shared request transport
 * (Go core bridge in the desktop app). Module-level so try/catch stays out
 * of the component. */
async function runQuery(
	sender: RequestSender,
	input: {
		endpoint: string;
		query: string;
		variables: string;
		headers: KeyValueRow[];
	},
): Promise<ResponseData> {
	const headers = input.headers.reduce<{ key: string; value: string }[]>((acc, h) => {
		if (h.enabled && h.key.trim() !== "") acc.push({ key: h.key, value: h.value });
		return acc;
	}, []);
	return sender({
		method: "POST",
		url: input.endpoint,
		headers,
		bodyType: "graphql",
		graphqlQuery: input.query,
		graphqlVariables: input.variables,
		sendId: crypto.randomUUID(),
	});
}

/** SchemaFieldRow renders one `name: Type` schema line; clicking navigates. */
function SchemaFieldRow({
	field,
	onNavigate,
}: {
	field: GqlField;
	onNavigate: (typeName: string) => void;
}) {
	const target = field.type?.of?.name ?? field.type?.name;
	const content = (
		<>
			<span className="font-data text-status-info">{field.name}</span>
			<span className="font-data text-muted-foreground">: {gqlTypeRef(field.type)}</span>
			{field.deprecated ? (
				<span className="ml-1 rounded border border-status-warn/40 px-1 font-data text-[10px] text-status-warn">
					deprecated
				</span>
			) : null}
		</>
	);
	return (
		<li className="border-b border-border/40 px-2 py-1 text-xs last:border-0">
			{target ? (
				<button
					type="button"
					className="text-left hover:text-primary"
					onClick={() => onNavigate(target)}
					title={`Go to ${target}`}
				>
					{content}
				</button>
			) : (
				content
			)}
		</li>
	);
}

/** TypeDetail is the schema-browser detail card for one selected type. */
function TypeDetail({
	typ,
	onNavigate,
}: {
	typ: GqlType;
	onNavigate: (typeName: string) => void;
}) {
	return (
		<div className="flex min-h-0 flex-1 flex-col overflow-y-auto rounded-xl border border-border bg-card">
			<p className="shrink-0 px-3 py-2 font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
				{typ.name} {typ.kind}
			</p>
			{typ.description ? (
				<p className="shrink-0 px-3 pb-2 text-[11px] text-muted-foreground">{typ.description}</p>
			) : null}
			{(typ.enumValues?.length ?? 0) > 0 ? (
				<p className="shrink-0 px-3 pb-2 font-data text-[11px] text-foreground">
					enum: {typ.enumValues!.join(" | ")}
				</p>
			) : null}
			<ul className="flex flex-col">
				{(typ.fields ?? []).map((f) => (
					<SchemaFieldRow key={`${typ.name}-${f.name}`} field={f} onNavigate={onNavigate} />
				))}
			</ul>
		</div>
	);
}

/** SchemaBrowser is the right-hand panel: filter, type chips, detail card. */
function SchemaBrowser({
	schema,
	busy,
	selected,
	onSelect,
	onIntrospect,
}: {
	schema: GqlSchema | null;
	busy: boolean;
	selected: string | null;
	onSelect: (name: string) => void;
	onIntrospect: () => void;
}) {
	const [query, setQuery] = useState("");
	const types = (schema?.types ?? []).filter((t) =>
		t.name.toLowerCase().includes(query.trim().toLowerCase()),
	);
	const selectedType = schema?.types?.find((t) => t.name === selected);
	return (
		<aside
			aria-label="Schema browser"
			className="flex w-80 shrink-0 flex-col gap-2 border-l border-border p-3"
		>
			<div className="flex items-center justify-between">
				<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
					Schema browser
				</p>
				<Button
					size="icon"
					variant="ghost"
					aria-label="Introspect schema"
					disabled={busy}
					onClick={onIntrospect}
				>
					{busy ? <Spinner className="size-3.5" /> : <RefreshCw className="size-3.5" />}
				</Button>
			</div>
			<Input
				value={query}
				onChange={(e) => setQuery(e.target.value)}
				placeholder="Filter types…"
				aria-label="Filter types"
				className="text-xs"
			/>
			<div className="flex flex-wrap gap-1.5">
				{types.map((t) => (
					<button
						key={t.name}
						type="button"
						onClick={() => onSelect(t.name)}
						className={cn(
							"flex items-center gap-1 rounded-full border px-2 py-0.5 font-data text-[10px]",
							selected === t.name
								? "border-primary/50 bg-primary/10 text-primary"
								: "border-border bg-muted/30 text-muted-foreground hover:text-foreground",
						)}
					>
						{t.name}
						<span className="opacity-60">{t.kind}</span>
					</button>
				))}
			</div>
			{selectedType ? (
				<TypeDetail typ={selectedType} onNavigate={onSelect} />
			) : (
				<div className="flex min-h-0 flex-1 items-center justify-center">
					<p className="max-w-44 text-center text-[11px] text-muted-foreground">
						{schema
							? "Pick a type to inspect its fields."
							: "Introspect an endpoint to browse its schema."}
					</p>
				</div>
			)}
		</aside>
	);
}

/** ResultArea renders the response pane: empty state, loading, or result. */
function ResultArea({
	loading,
	response,
	error,
}: {
	loading: boolean;
	response: ResponseData | null;
	error: string | null;
}) {
	if (loading) {
		return (
			<div
				className="flex min-h-0 flex-1 flex-col items-center justify-center gap-3"
				role="status"
				aria-label="Running operation"
			>
				<Spinner className="size-6" />
				<p className="text-xs text-muted-foreground">Running…</p>
			</div>
		);
	}
	if (error) {
		return (
			<div className="min-h-0 flex-1 overflow-y-auto rounded-xl border border-border bg-background p-2">
				<CodeMirrorEditor value={`// Error: ${error}`} language="text" readOnly className="h-full" />
			</div>
		);
	}
	if (response) {
		const parsed = parseJsonBody(response.body);
		return (
			<div className="flex min-h-0 flex-1 flex-col gap-1">
				<div className="flex shrink-0 items-center gap-2 font-data text-xs tabular-nums text-muted-foreground">
					<StatusPill status={response.statusCode} />
					<span>{response.durationMs}ms</span>
					<span aria-hidden>·</span>
					<span>{formatBytes(response.size)}</span>
				</div>
				<div className="min-h-0 flex-1 overflow-y-auto rounded-xl border border-border bg-background p-2">
					{parsed !== null ? (
						<JsonTree data={parsed} />
					) : (
						<CodeMirrorEditor
							value={prettyBody(response.body, "")}
							language="text"
							readOnly
							className="h-full"
						/>
					)}
				</div>
			</div>
		);
	}
	return (
		<div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-1.5">
			<Settings className="size-8 text-muted-foreground/40" aria-hidden />
			<p className="text-sm font-medium text-foreground">No result yet</p>
			<p className="text-xs text-muted-foreground">Run the operation to see data here.</p>
		</div>
	);
}

/** GraphqlBrowser is the G-17.4.2 GraphQL playground: endpoint bar with Run,
 * Query/Variables/Headers tabs, results pane, and a schema-browser side
 * panel fed by bridge introspection. Queries ride the shared Go-core
 * transport via the request store's sender. */
export function GraphqlBrowser() {
	const [endpoint, setEndpoint] = useState("");
	const [authHeader, setAuthHeader] = useState("");
	const [tab, setTab] = useState<EditorTab>("query");
	const [query, setQuery] = useState("");
	const [variables, setVariables] = useState("{\n  \n}");
	const [headerRows, setHeaderRows] = useState<KeyValueRow[]>([
		{ key: "Authorization", value: "", enabled: true },
	]);
	const [schema, setSchema] = useState<GqlSchema | null>(null);
	const [introspecting, setIntrospecting] = useState(false);
	const [running, setRunning] = useState(false);
	const [response, setResponse] = useState<ResponseData | null>(null);
	const [runError, setRunError] = useState<string | null>(null);
	const [introspectError, setIntrospectError] = useState<string | null>(null);
	const [selectedType, setSelectedType] = useState<string | null>(null);

	const operations = parseOperations(query);

	const introspect = (): void => {
		if (endpoint.trim() === "") return;
		setIntrospecting(true);
		setIntrospectError(null);
		const headers =
			authHeader.trim() === ""
				? undefined
				: [{ key: "Authorization", value: `Bearer ${authHeader.trim()}` }];
		getGqlBridge()
			.introspect({ endpoint: endpoint.trim(), headers, timeoutSec: 30 })
			.then((s) => {
				setSchema(s);
				setIntrospecting(false);
			})
			.catch((error) => {
				setIntrospectError(error instanceof Error ? error.message : String(error));
				setIntrospecting(false);
			});
	};

	const run = (): void => {
		if (endpoint.trim() === "" || query.trim() === "") return;
		setRunning(true);
		setRunError(null);
		const rows: KeyValueRow[] =
			authHeader.trim() === ""
				? headerRows
				: [
						...headerRows.filter((h) => h.key.toLowerCase() !== "authorization"),
						{ key: "Authorization", value: `Bearer ${authHeader.trim()}`, enabled: true },
					];
		runQuery(useRequestStore.getState().sender, {
			endpoint: endpoint.trim(),
			query,
			variables,
			headers: rows,
		})
			.then(setResponse)
			.catch((error) => {
				setRunError(error instanceof Error ? error.message : String(error));
			})
			.finally(() => {
				setRunning(false);
			});
	};

	return (
		<div className="flex h-full min-h-0">
			<section
				aria-label="GraphQL client"
				className="flex min-w-0 flex-1 flex-col gap-2 p-4"
			>
				<div className="flex items-center gap-2">
					<span
						aria-hidden
						className="shrink-0 rounded-full border border-border bg-muted/40 px-2 py-0.5 font-data text-[10px] font-semibold"
					>
						GQL
					</span>
					<Input
						value={endpoint}
						onChange={(e) => setEndpoint(e.target.value)}
						placeholder="{{graphqlUrl}}"
						aria-label="GraphQL endpoint"
						spellCheck={false}
						className="flex-1 font-mono text-xs"
					/>
					<select
						aria-label="Operation"
						className="h-8 w-28 shrink-0 rounded-md border border-border bg-transparent px-2 font-data text-xs"
						defaultValue=""
					>
						<option value="">&lt;auto&gt;</option>
						{operations.map((op) => (
							<option key={`${op.type}-${op.name}`} value={op.name}>
								{op.name}
							</option>
						))}
					</select>
					<Button
						size="icon"
						variant="ghost"
						aria-label="Introspect schema"
						disabled={introspecting || endpoint.trim() === ""}
						onClick={introspect}
					>
						{introspecting ? (
							<Spinner className="size-3.5" />
						) : (
							<RefreshCw className="size-3.5" />
						)}
					</Button>
					<Button
						size="sm"
						variant="destructive"
						disabled={running || endpoint.trim() === "" || query.trim() === ""}
						onClick={() => run()}
					>
						{running ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
						Run
					</Button>
				</div>

				<div className="flex shrink-0 items-center gap-1">
					<Tabs
						value={tab}
						onValueChange={(v) => {
							// SAFETY: tab ids come from the local TABS constant
							setTab(v as EditorTab)
						}}
					>
						<TabsList variant="line" aria-label="GraphQL editor tabs">
							{TABS.map((t) => (
								<TabsTrigger key={t.id} value={t.id} className={tabClass(tab === t.id)}>
									{t.label}
								</TabsTrigger>
							))}
						</TabsList>
					</Tabs>
					<button
						type="button"
						onClick={() => setQuery((q) => prettifyQuery(q))}
						className="rounded-full border border-border px-2.5 py-0.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
					>
						Prettify
					</button>
				</div>

				<div className="h-64 shrink-0 overflow-hidden rounded-xl border border-border">
					{tab === "query" ? (
						<CodeMirrorEditor
							value={query}
							onChange={setQuery}
							language="graphql"
							className="h-full"
						/>
					) : tab === "variables" ? (
						<CodeMirrorEditor
							value={variables}
							onChange={setVariables}
							language="json"
							className="h-full"
						/>
					) : (
						<div className="h-full overflow-y-auto bg-background p-2">
							<KeyValueEditor
								rows={headerRows}
								onChange={setHeaderRows}
								keyPlaceholder="Header"
								valuePlaceholder="Value"
							/>
							<div className="mt-2 flex items-center gap-2">
								<span className="text-[11px] text-muted-foreground">Bearer token</span>
								<Input
									value={authHeader}
									onChange={(e) => setAuthHeader(e.target.value)}
									placeholder="eyJ…"
									type="password"
									aria-label="Bearer token"
									spellCheck={false}
									className="h-7 flex-1 font-mono text-xs"
								/>
							</div>
						</div>
					)}
				</div>

				{(introspectError ?? runError) ? (
					<Alert variant="destructive">
						<AlertDescription>{introspectError ?? runError}</AlertDescription>
					</Alert>
				) : null}

				<ResultArea loading={running} response={response} error={runError} />
			</section>

			<SchemaBrowser
				schema={schema}
				busy={introspecting}
				selected={selectedType}
				onSelect={setSelectedType}
				onIntrospect={introspect}
			/>
		</div>
	);
}
