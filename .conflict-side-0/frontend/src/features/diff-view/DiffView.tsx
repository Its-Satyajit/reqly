import { useEffect, useState } from "react";
import { ArrowUpDown, GitCompareArrows, TriangleAlert } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import {
	changeLabel,
	getDiffBridge,
	type DiffAdapter,
	type DiffChange,
	type DiffResultView,
	type ResponseDiffResult,
	type SpecDiffResult,
} from "#lib/diff";
import type { HistoryEntry } from "#lib/history";
import { methodTintClass } from "#lib/status";
import { useHistoryStore } from "#stores/useHistoryStore";

type DiffMode = "specs" | "responses";

type ChangeKind = "added" | "removed" | "changed";

/** kindOf maps a diff change type onto the reference's four count cards. */
function kindOf(c: DiffChange): ChangeKind {
	if (c.type === "create") return "added";
	if (c.type === "delete") return "removed";
	return "changed";
}

const KIND_PILL = {
	added: "border-status-ok/40 text-status-ok",
	removed: "border-status-error/40 text-status-error",
	changed: "border-status-warn/40 text-status-warn",
} satisfies Record<ChangeKind, string>;

const KIND_DOT = {
	added: "bg-status-ok",
	removed: "bg-status-error",
	changed: "bg-status-warn",
} satisfies Record<ChangeKind, string>;

/** changeTitle renders the reference's bold first line: the changed element
 * path, joined with → for nesting. */
function changeTitle(c: DiffChange): string {
	return c.path.join(" → ") || "(root)";
}

/** changeSummary is the muted second line under the title. */
function changeSummary(c: DiffChange): string {
	if (c.severity === "breaking") {
		return `${kindOf(c)} — breaking change`;
	}
	return changeLabel(c);
}

/** matchesFilter reports whether a change survives the filter input. */
function matchesFilter(c: DiffChange, query: string): boolean {
	if (query.trim() === "") return true;
	const hay = `${changeTitle(c)} ${changeSummary(c)} ${kindOf(c)} ${c.severity ?? ""}`;
	return hay.toLowerCase().includes(query.trim().toLowerCase());
}

/** sortChanges puts breaking first, then removals, additions, changes. */
function sortChanges(changes: DiffChange[]): DiffChange[] {
	const rank = (c: DiffChange): number => {
		if (c.severity === "breaking") return 0;
		const order = { removed: 1, added: 2, changed: 3 } as const;
		return order[kindOf(c)];
	};
	return [...changes].sort((a, b) => rank(a) - rank(b));
}

/** CountCard is one reference stat tile (added / removed / changed / BREAKING). */
function CountCard({
	count,
	label,
	tone,
}: {
	count: number;
	label: string;
	tone: "ok" | "error" | "warn" | "breaking";
}) {
	const toneClass = {
		ok: "text-status-ok",
		error: "text-status-error",
		warn: "text-status-warn",
		breaking: "text-status-error border-status-error/50",
	}[tone];
	return (
		<div
			className={cn(
				"flex min-w-20 flex-col items-center gap-0.5 rounded-xl border border-border bg-card px-3 py-2",
				tone === "breaking" && count > 0 && toneClass,
			)}
		>
			<span className={cn("font-data text-xl font-semibold", tone !== "breaking" && toneClass)}>
				{count}
			</span>
			<span className="text-[11px] uppercase tracking-wide text-muted-foreground">{label}</span>
		</div>
	);
}

/** ChangeRow is one selectable change line: severity dot, title, summary,
 * and a right-aligned kind pill. */
function ChangeRow({
	change,
	selected,
	onSelect,
}: {
	change: DiffChange;
	selected: boolean;
	onSelect: () => void;
}) {
	const kind = kindOf(change);
	return (
		<button
			type="button"
			onClick={onSelect}
			className={cn(
				"flex w-full items-start gap-2 rounded-lg border px-2.5 py-2 text-left transition-colors",
				selected
					? "border-primary/50 bg-primary/5"
					: "border-transparent hover:bg-accent",
			)}
		>
			<span className={cn("mt-1.5 size-2 shrink-0 rounded-full", KIND_DOT[kind])} aria-hidden />
			<span className="min-w-0 flex-1">
				<span className="block truncate text-xs font-semibold text-foreground">
					{changeTitle(change)}
				</span>
				<span className="block truncate text-[11px] text-muted-foreground">
					{changeSummary(change)}
				</span>
			</span>
			<span
				className={cn(
					"shrink-0 rounded-full border px-2 py-px font-data text-[10px] lowercase",
					KIND_PILL[kind],
				)}
			>
				{kind}
			</span>
		</button>
	);
}

/** DetailPane shows the selected change: title, severity badge, from/to, and
 * the breaking warning banner. */
function DetailPane({ change }: { change: DiffChange }) {
	const breaking = change.severity === "breaking";
	return (
		<div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto rounded-xl border border-border bg-card p-4">
			<h3 className="text-sm font-semibold text-foreground">
				Detail · {changeTitle(change)}
			</h3>
			<div className="flex items-center gap-2">
				<span
					className={cn(
						"rounded-full border px-2 py-0.5 font-data text-[10px] lowercase",
						breaking
							? "border-status-error/40 text-status-error"
							: "border-border text-muted-foreground",
					)}
				>
					severity: {change.severity ?? kindOf(change)}
				</span>
			</div>
			<p className="text-xs text-muted-foreground">{changeSummary(change)}</p>
			<div className="rounded-lg border border-border bg-muted/20 p-2 font-mono text-[11px]">
				<p className="text-muted-foreground">~ {changeLabel(change)}</p>
				<pre className="max-h-48 overflow-auto whitespace-pre-wrap break-all text-foreground">
					{`from: ${change.from === undefined ? "∅" : JSON.stringify(change.from, null, 2)}`}
					{`\nto:   ${change.to === undefined ? "∅" : JSON.stringify(change.to, null, 2)}`}
				</pre>
			</div>
			{breaking ? (
				<div className="flex items-start gap-2 rounded-lg border border-status-error/40 bg-status-error/10 px-3 py-2 text-xs text-status-error">
					<TriangleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden />
					<span>
						<strong>Breaking.</strong> Consumers relying on this contract will
						fail after deploy. Gate this release behind a major version bump.
					</span>
				</div>
			) : null}
		</div>
	);
}

/** ChangesPane is the left column: count cards, filter, and the change list. */
function ChangesPane({ result }: { result: DiffResultView }) {
	const [query, setQuery] = useState("");
	const [selectedKey, setSelectedKey] = useState<string | null>(null);
	const changes = sortChanges(result.changes ?? []);
	const filtered = changes.filter((c) => matchesFilter(c, query));
	const counts = {
		added: changes.filter((c) => kindOf(c) === "added").length,
		removed: changes.filter((c) => kindOf(c) === "removed").length,
		changed: changes.filter((c) => kindOf(c) === "changed").length,
		breaking: changes.filter((c) => c.severity === "breaking").length,
	};
	const selected =
		changes.find((c) => changeKey(c) === selectedKey) ?? filtered[0] ?? null;
	return (
		<div className="flex min-h-0 flex-1 gap-3">
			<div className="flex min-w-0 flex-1 flex-col gap-3 rounded-xl border border-border bg-card p-3">
				<div className="flex flex-wrap items-center gap-2">
					<CountCard count={counts.added} label="added" tone="ok" />
					<CountCard count={counts.removed} label="removed" tone="error" />
					<CountCard count={counts.changed} label="changed" tone="warn" />
					<CountCard count={counts.breaking} label="breaking" tone="breaking" />
					<Input
						value={query}
						onChange={(e) => setQuery(e.target.value)}
						placeholder="Filter changes…"
						aria-label="Filter changes"
						className="ml-auto w-44 text-xs"
					/>
				</div>
				{!result.hasChanges ? (
					<p className="rounded-lg border border-border px-3 py-2 text-xs text-status-ok">
						No differences.
					</p>
				) : (
					<div className="min-h-0 flex-1 overflow-y-auto">
						{filtered.length === 0 ? (
							<p className="px-2 py-1 text-xs text-muted-foreground">
								No changes match your filter.
							</p>
						) : (
							<ul className="flex flex-col gap-0.5">
								{filtered.map((c) => {
									const key = changeKey(c);
									return (
										<li key={key}>
											<ChangeRow
												change={c}
												selected={selected != null && key === changeKey(selected)}
												onSelect={() => setSelectedKey(key)}
											/>
										</li>
									);
								})}
							</ul>
						)}
					</div>
				)}
			</div>
			{selected ? <DetailPane change={selected} /> : null}
		</div>
	);
}

/** changeKey builds a stable key from the full change identity. */
function changeKey(c: DiffChange): string {
	return `${c.type}-${c.path.join("/")}-${JSON.stringify(c.from)}-${JSON.stringify(c.to)}`;
}

export function DiffView({ adapter }: { adapter?: DiffAdapter }) {
	const effective = adapter ?? getDiffBridge();
	const [mode, setMode] = useState<DiffMode>("specs");
	const [pathA, setPathA] = useState("");
	const [pathB, setPathB] = useState("");
	const [entryAId, setEntryAId] = useState("");
	const [entryBId, setEntryBId] = useState("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [specResult, setSpecResult] = useState<SpecDiffResult | null>(null);
	const [respResult, setRespResult] = useState<ResponseDiffResult | null>(null);

	// Pull from the shared history store so state updates live there, not in
	// this component's effect.
	useEffect(() => {
		if (mode === "responses") {
			void useHistoryStore.getState().loadPool();
		}
	}, [mode]);
	const entries = useHistoryStore((s) => s.pool);

	// Promise-chain form keeps hook updates out of try/catch, which the React
	// Compiler cannot model.
	const run = (): void => {
		if (mode === "specs" && (pathA.trim() === "" || pathB.trim() === "")) return;
		if (mode === "responses" && (entryAId === "" || entryBId === "")) return;
		setBusy(true);
		setError(null);
		const pending =
			mode === "specs"
				? effective.specs(pathA.trim(), pathB.trim()).then(setSpecResult)
				: effective.responses(entryAId, entryBId).then(setRespResult);
		pending
			.catch((error) => {
				setError(error instanceof Error ? error.message : String(error));
			})
			.finally(() => {
				setBusy(false);
			});
	};

	const swap = (): void => {
		if (mode === "specs") {
			setPathA(pathB);
			setPathB(pathA);
		} else {
			setEntryAId(entryBId);
			setEntryBId(entryAId);
		}
	};

	const entryOption = (e: HistoryEntry | null) =>
		e
			? `${e.method} ${e.url} · ${e.status} · ${new Date(e.createdAt).toLocaleString()}`
			: "";

	const result = mode === "specs" ? specResult?.result ?? null : respResult?.result ?? null;

	return (
		<section className="flex h-full min-h-0 flex-col gap-3 p-4" aria-label="API diff">
			<div className="flex flex-wrap items-center gap-2">
				<span className="text-xs text-muted-foreground">Base</span>
				{mode === "specs" ? (
					<Input
						aria-label="Base spec"
						value={pathA}
						onChange={(e) => setPathA(e.target.value)}
						placeholder="specs/old.yaml"
						spellCheck={false}
						className="w-56 font-mono text-xs"
					/>
				) : (
					<select
						aria-label="Base entry"
						value={entryAId}
						onChange={(e) => setEntryAId(e.target.value)}
						className="h-8 w-56 rounded-md border border-border bg-transparent px-2 text-xs"
					>
						<option value="">Pick entry…</option>
						{entries.map((e) => (
							<option key={e.id} value={e.id}>
								{entryOption(e)}
							</option>
						))}
					</select>
				)}
				<Button
					size="icon"
					variant="ghost"
					aria-label="Swap base and updated"
					onClick={swap}
				>
					<ArrowUpDown className="size-3.5" aria-hidden />
				</Button>
				<span className="text-xs text-muted-foreground">Updated</span>
				{mode === "specs" ? (
					<Input
						aria-label="Updated spec"
						value={pathB}
						onChange={(e) => setPathB(e.target.value)}
						placeholder="specs/new.yaml"
						spellCheck={false}
						className="w-56 font-mono text-xs"
					/>
				) : (
					<select
						aria-label="Updated entry"
						value={entryBId}
						onChange={(e) => setEntryBId(e.target.value)}
						className="h-8 w-56 rounded-md border border-border bg-transparent px-2 text-xs"
					>
						<option value="">Pick entry…</option>
						{entries.map((e) => (
							<option key={e.id} value={e.id}>
								{entryOption(e)}
							</option>
						))}
					</select>
				)}
				<Button
					size="sm"
					variant="destructive"
					disabled={
						busy ||
						(mode === "specs" && (pathA.trim() === "" || pathB.trim() === "")) ||
						(mode === "responses" &&
							(entryAId === "" || entryBId === "" || entryAId === entryBId))
					}
					onClick={() => run()}
				>
					{busy ? (
						<Spinner data-icon="inline-start" />
					) : (
						<GitCompareArrows data-icon="inline-start" />
					)}
					Compare
				</Button>
				<Button
					size="sm"
					variant="outline"
					className="ml-auto"
					onClick={() => setMode(mode === "specs" ? "responses" : "specs")}
				>
					Response diff (from history)
				</Button>
			</div>

			{error ? (
				<Alert variant="destructive">
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			) : null}

			{mode === "responses" && respResult?.metaA && respResult.metaB ? (
				<div className="grid shrink-0 grid-cols-2 gap-2 text-[11px]">
					{[respResult.metaA, respResult.metaB].map((m) => (
						<div
							key={m.id}
							className="rounded-lg border border-border bg-card px-2 py-1.5 font-mono"
						>
							<p className="flex items-center gap-1.5">
								<span className={cn("font-semibold uppercase", methodTintClass(m.method))}>
									{m.method}
								</span>
								<span className={m.status >= 400 ? "text-status-error" : "text-status-ok"}>
									{m.status}
								</span>
							</p>
							<p className="truncate text-muted-foreground">{m.url}</p>
							{m.env !== "" && (
								<p className="text-muted-foreground">env: {m.env}</p>
							)}
						</div>
					))}
				</div>
			) : null}

			{result ? <ChangesPane result={result} /> : (
				<div className="flex min-h-0 flex-1 items-center justify-center rounded-xl border border-border bg-card">
					<p className="text-xs text-muted-foreground">
						Pick a base and an updated {mode === "specs" ? "spec" : "pair of responses"}, then
						hit Compare.
					</p>
				</div>
			)}
		</section>
	);
}
