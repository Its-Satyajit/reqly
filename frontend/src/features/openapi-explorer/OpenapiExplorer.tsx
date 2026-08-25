import { useState } from "react";
import { FolderOutput, RefreshCw } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import { methodTintClass } from "#lib/status";
import {
	getOpenapiBridge,
	type OpenapiAdapter,
	type OpenapiEndpointView,
} from "#lib/openapi";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

const METHOD_FILTERS = ["GET", "POST", "PUT", "PATCH", "DELETE"] as const;

function FilterChip({
	label,
	active,
	onClick,
}: {
	label: string;
	active: boolean;
	onClick: () => void;
}) {
	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"rounded-full border px-2.5 py-0.5 text-xs transition-colors",
				active
					? "border-primary/50 bg-primary/10 font-medium text-primary"
					: "border-border bg-muted/30 text-muted-foreground hover:text-foreground",
			)}
		>
			{label}
		</button>
	);
}

/** EndpointRow is one selectable tree line: method chip, path, summary. */
function EndpointRow({
	endpoint,
	selected,
	checked,
	onOpen,
	onToggle,
}: {
	endpoint: OpenapiEndpointView;
	selected: boolean;
	checked: boolean;
	onOpen: () => void;
	onToggle: () => void;
}) {
	return (
		<div
			className={cn(
				"flex items-start gap-2 rounded-lg px-2 py-1.5",
				selected ? "bg-primary/5" : "hover:bg-accent/50",
			)}
		>
			<input
				type="checkbox"
				checked={checked}
				onChange={onToggle}
				aria-label={`Select ${endpoint.method} ${endpoint.path} for generation`}
				className="mt-0.5 size-3.5 shrink-0 accent-(--primary)"
			/>
			<button type="button" className="min-w-0 flex-1 text-left" onClick={onOpen}>
				<span className="flex items-center gap-2">
					<span
						className={cn(
							"shrink-0 rounded-full border border-border bg-muted/40 px-1.5 py-px font-data text-2xs font-semibold uppercase",
							methodTintClass(endpoint.method),
						)}
					>
						{endpoint.method}
					</span>
					<span className="truncate font-mono text-xs text-foreground">{endpoint.path}</span>
				</span>
				{endpoint.summary ? (
					<span className="block truncate pl-1 text-xs text-muted-foreground">
						{endpoint.summary}
					</span>
				) : null}
			</button>
		</div>
	);
}

/** SchemaBlock is one labelled preformatted schema. */
function SchemaBlock({ label, schema }: { label: string; schema: string }) {
	if (schema === "") return null;
	return (
		<div className="flex flex-col gap-1">
			<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
				{label}
			</p>
			<pre className="max-h-56 overflow-auto whitespace-pre-wrap break-all rounded-lg border border-border bg-background p-2 font-mono text-xs">
				{schema}
			</pre>
		</div>
	);
}

/** EndpointDetail is the right pane for the selected operation. */
function EndpointDetail({ endpoint }: { endpoint: OpenapiEndpointView }) {
	const statuses = Object.keys(endpoint.responseSchemas ?? {});
	return (
		<div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto rounded-xl border border-border bg-card p-4">
			<div className="flex flex-wrap items-center gap-2">
				<span
					className={cn(
						"shrink-0 rounded-full border border-border bg-muted/40 px-2 py-0.5 font-data text-2xs font-semibold uppercase",
						methodTintClass(endpoint.method),
					)}
				>
					{endpoint.method}
				</span>
				<h3 className="truncate font-mono text-sm font-semibold text-foreground">
					{endpoint.path}
				</h3>
			</div>
			{endpoint.operationId ? (
				<p className="font-data text-xs text-muted-foreground">{endpoint.operationId}</p>
			) : null}
			{endpoint.summary ? <p className="text-xs text-foreground">{endpoint.summary}</p> : null}
			<SchemaBlock label="Request body" schema={endpoint.requestSchema ?? ""} />
			{statuses.map((status) => (
				<SchemaBlock
					key={status}
					label={`Response ${status}`}
					schema={endpoint.responseSchemas?.[status] ?? ""}
				/>
			))}
		</div>
	);
}

type ExplorerState = {
	specPath: string;
	result: Awaited<ReturnType<OpenapiAdapter["explore"]>> | null;
	selected: string[];
	dirName: string;
	busy: boolean;
	error: string | null;
	generated: string[] | null;
};

/** OpenapiExplorer (G-17.4.8): spec tree grouped by tag with search +
 * tag/method filter chips, an endpoint detail pane, and request generation
 * from selected operations. Try-it / security schemes await a bridge seam. */
export function OpenapiExplorer() {
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const [ui, setUi] = useState<ExplorerState>({
		specPath: "",
		result: null,
		selected: [],
		dirName: "",
		busy: false,
		error: null,
		generated: null,
	});
	const patch = (p: Partial<ExplorerState>) => setUi((prev) => ({ ...prev, ...p }));
	const { specPath, result, selected, dirName, busy, error, generated } = ui;
	const [search, setSearch] = useState("");
	const [tagFilter, setTagFilter] = useState<string | null>(null);
	const [methodFilter, setMethodFilter] = useState<string | null>(null);
	const [detailKey, setDetailKey] = useState<string | null>(null);

	const selectedSet = new Set(selected);
	const detail =
		result?.endpoints.find((ep) => `${ep.method}|${ep.path}` === detailKey) ?? null;

	const tags = [...new Set((result?.endpoints ?? []).map((ep) => ep.tags?.[0] ?? "untagged"))].sort();
	// SAFETY: METHOD_FILTERS is a readonly string tuple; includes() narrows by value.
	const knownMethods: readonly string[] = METHOD_FILTERS;
	const presentMethods = [
		...new Set((result?.endpoints ?? []).map((ep) => ep.method)),
	].filter((m) => knownMethods.includes(m));

	const matches = (ep: OpenapiEndpointView): boolean => {
		if (tagFilter != null && (ep.tags?.[0] ?? "untagged") !== tagFilter) return false;
		if (methodFilter != null && ep.method !== methodFilter) return false;
		if (search.trim() !== "") {
			const hay = `${ep.path} ${ep.summary ?? ""} ${ep.operationId ?? ""}`.toLowerCase();
			if (!hay.includes(search.trim().toLowerCase())) return false;
		}
		return true;
	};

	const grouped: [string, OpenapiEndpointView[]][] = [];
	{
		const byTag = new Map<string, OpenapiEndpointView[]>();
		for (const ep of result?.endpoints ?? []) {
			if (!matches(ep)) continue;
			const tag = ep.tags?.[0] ?? "untagged";
			const list = byTag.get(tag) ?? [];
			list.push(ep);
			byTag.set(tag, list);
		}
		grouped.push(...[...byTag.entries()].sort(([a], [b]) => a.localeCompare(b)));
	}

	const explore = (): void => {
		if (specPath.trim() === "") return;
		patch({ busy: true, error: null, result: null, selected: [], generated: null });
		setDetailKey(null);
		getOpenapiBridge()
			.explore(specPath.trim())
			.then((res) => {
				patch({ result: res, busy: false });
			})
			.catch((e) => {
				patch({ error: e instanceof Error ? e.message : String(e), busy: false });
			});
	};

	const toggle = (method: string, path: string): void => {
		const key = `${method}|${path}`;
		patch({
			selected: selected.includes(key)
				? selected.filter((k) => k !== key)
				: [...selected, key],
		});
	};

	const generate = (): void => {
		if (specPath.trim() === "" || selected.length === 0 || dirName.trim() === "") return;
		patch({ busy: true, error: null });
		const selections = selected.map((k) => {
			// SAFETY: keys are built above as `METHOD|path` pairs.
			const idx = k.indexOf("|");
			return { method: k.slice(0, idx), path: k.slice(idx + 1) };
		});
		getOpenapiBridge()
			.generate({ specPath: specPath.trim(), selections, dirName: dirName.trim() })
			.then((res) => {
				patch({ generated: res.created, busy: false });
				void refreshWorkspace();
			})
			.catch((e) => {
				patch({ error: e instanceof Error ? e.message : String(e), busy: false });
			});
	};

	return (
		<div className="flex h-full min-h-0 gap-4 p-4" aria-label="OpenAPI explorer">
			<section className="flex w-96 shrink-0 flex-col gap-2 overflow-y-auto">
				<div className="flex items-end gap-2">
					<Input
						value={specPath}
						onChange={(e) => patch({ specPath: e.target.value })}
						placeholder="specs/pets.yaml"
						spellCheck={false}
						aria-label="Spec path (workspace-relative)"
						className="flex-1 font-mono text-xs"
					/>
					<Button
						size="sm"
						variant="outline"
						disabled={busy || specPath.trim() === ""}
						onClick={explore}
					>
						{busy ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
						Explore
					</Button>
				</div>

				{error ? (
					<Alert variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				) : null}

				{result ? (
					<>
						<p className="text-xs text-muted-foreground">
							<span className="font-medium text-foreground">{result.title}</span>
							{result.version != null && result.version !== "" ? ` · v${result.version}` : ""}
							{` · ${result.endpoints.length} operations`}
						</p>
						<Input
							value={search}
							onChange={(e) => setSearch(e.target.value)}
							placeholder="Search endpoints…"
							aria-label="Search endpoints"
							className="text-xs"
						/>
						<div className="flex flex-wrap gap-1.5">
							<FilterChip
								label="all tags"
								active={tagFilter == null}
								onClick={() => setTagFilter(null)}
							/>
							{tags.map((t) => (
								<FilterChip
									key={t}
									label={t}
									active={tagFilter === t}
									onClick={() => setTagFilter(tagFilter === t ? null : t)}
								/>
							))}
						</div>
						<div className="flex flex-wrap gap-1.5">
							<FilterChip
								label="all methods"
								active={methodFilter == null}
								onClick={() => setMethodFilter(null)}
							/>
							{presentMethods.map((m) => (
								<FilterChip
									key={m}
									label={m}
									active={methodFilter === m}
									onClick={() => setMethodFilter(methodFilter === m ? null : m)}
								/>
							))}
						</div>

						{grouped.map(([tag, eps]) => (
							<div key={tag} className="flex flex-col gap-0.5">
								<p className="px-1 pt-1 font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
									{tag}
								</p>
								{eps.map((ep) => {
									const key = `${ep.method}|${ep.path}`;
									return (
										<EndpointRow
											key={key}
											endpoint={ep}
											selected={detailKey === key}
											checked={selectedSet.has(key)}
											onOpen={() => setDetailKey(key)}
											onToggle={() => toggle(ep.method, ep.path)}
										/>
									);
								})}
							</div>
						))}

						<div className="flex items-center gap-2 border-t border-border pt-2">
							<Input
								value={dirName}
								onChange={(e) => patch({ dirName: e.target.value })}
								placeholder="collections/<name>"
								spellCheck={false}
								aria-label="Target collection directory"
								className="flex-1 font-mono text-xs"
							/>
							<Button
								size="sm"
								variant="destructive"
								disabled={busy || selected.length === 0 || dirName.trim() === ""}
								onClick={generate}
							>
								<FolderOutput data-icon="inline-start" />
								Generate
							</Button>
						</div>
						{generated ? (
							<p className="text-xs text-status-ok">
								Created {generated.length} request file(s) — see the sidebar.
							</p>
						) : null}
					</>
				) : (
					<p className="text-xs text-muted-foreground">
						Point at a workspace-relative OpenAPI spec and hit Explore.
					</p>
				)}
			</section>

			{detail ? (
				<EndpointDetail endpoint={detail} />
			) : (
				<div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-1.5 rounded-xl border border-border bg-card">
					<p className="text-sm font-medium text-foreground">Select an endpoint</p>
					<p className="max-w-xs text-center text-xs text-muted-foreground">
						Browse the specification tree — every operation carries params and
						schemas.
					</p>
				</div>
			)}
		</div>
	);
}
