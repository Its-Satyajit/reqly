import { useEffect, useState } from "react";
import { GitCompareArrows } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CompactSelect } from "#components/CompactSelect";
import { cn } from "#lib/utils";
import {
  changeLabel,
  getDiffBridge,
  type DiffAdapter,
  type DiffChange,
  type DiffResultView,
  type ResponseDiffResult,
} from "#lib/diff";
import type { HistoryEntry } from "#lib/history";
import { useHistoryStore } from "#stores/useHistoryStore";

type DiffMode = "specs" | "responses";

const MODE_OPTIONS = [
  { value: "specs", label: "OpenAPI specs" },
  { value: "responses", label: "History responses" },
];

function severityBadgeClass(severity?: string): string {
  switch (severity) {
    case "breaking":
      return "text-status-error border-status-error/40";
    case "addition":
      return "text-status-ok border-status-ok/40";
    default:
      return "";
  }
}

function ChangesList({ result, severity }: { result: DiffResultView; severity: boolean }) {
  if (!result.hasChanges) {
    return (
      <p className="rounded-md border border-border px-3 py-2 text-xs text-status-ok">
        No differences.
      </p>
    );
  }
  const changes = [...(result.changes ?? [])];
  // Breaking changes first so the review starts with the risky ones.
  changes.sort((a, b) => rank(a.severity) - rank(b.severity));
  return (
    <ul className="flex flex-col gap-1">
      {changes.map((c, i) => (
        <ChangeRow key={changeKey(c, i)} change={c} severity={severity} />
      ))}
    </ul>
  );
}

function rank(severity?: string): number {
  if (severity === "breaking") return 0;
  if (severity === "addition") return 1;
  return severity === undefined ? 2 : 2;
}

/** changeKey builds a stable key from the change identity (no list index). */
function changeKey(c: DiffChange, fallback: number): string {
  return `${c.type}-${c.path.join("/")}-${fallback}`;
}

function ChangeRow({ change, severity }: { change: DiffChange; severity: boolean }) {
  const [expanded, setExpanded] = useState(false);
  return (
    <li className="flex flex-col gap-0.5 rounded-md border border-border px-2 py-1.5 text-xs">
      <button
        type="button"
        className="flex items-center gap-2 text-left"
        onClick={() => setExpanded(!expanded)}
      >
        <Badge variant="outline" className={cn("shrink-0", severityBadgeClass(change.severity))}>
          {severity ? (change.severity ?? change.type) : change.type}
        </Badge>
        <span className="truncate font-mono">{changeLabel(change)}</span>
      </button>
      {expanded && (
        <div className="grid grid-cols-2 gap-2 pl-6 font-mono text-[11px]">
          <div className="min-w-0">
            <p className="pb-0.5 font-sans text-muted-foreground">from</p>
            <pre className="max-h-40 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/50 p-1">
              {JSON.stringify(change.from) ?? "∅"}
            </pre>
          </div>
          <div className="min-w-0">
            <p className="pb-0.5 font-sans text-muted-foreground">to</p>
            <pre className="max-h-40 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/50 p-1">
              {JSON.stringify(change.to) ?? "∅"}
            </pre>
          </div>
        </div>
      )}
    </li>
  );
}

export function DiffView({ adapter }: { adapter?: DiffAdapter }) {
  const effective = adapter ?? getDiffBridge();
  const [mode, setMode] = useState<DiffMode>("specs");
  const [pathA, setPathA] = useState("");
  const [pathB, setPathB] = useState("");
  const [entryA, setEntryA] = useState<HistoryEntry | null>(null);
  const [entryB, setEntryB] = useState<HistoryEntry | null>(null);
  const [entries, setEntries] = useState<HistoryEntry[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [specResult, setSpecResult] = useState<Awaited<ReturnType<DiffAdapter["specs"]>> | null>(null);
  const [respResult, setRespResult] = useState<ResponseDiffResult | null>(null);

  const loadEntries = async () => {
    try {
      const list = await useHistoryStore.getState().adapter.list(100, 0, "", "");
      setEntries(list);
    } catch {
      /* history unavailable; pickers stay empty */
    }
  };

  useEffect(() => {
    if (mode === "responses") void loadEntries();
  }, [mode]);

  const run = async () => {
    setBusy(true);
    setError(null);
    try {
      if (mode === "specs") {
        if (pathA.trim() === "" || pathB.trim() === "") return;
        setSpecResult(await effective.specs(pathA.trim(), pathB.trim()));
      } else if (mode === "responses") {
        if (!entryA || !entryB) return;
        setRespResult(await effective.responses(entryA.id, entryB.id));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  const entryOption = (e: HistoryEntry | null) =>
    e ? `${e.method} ${e.url} · ${e.status} · ${new Date(e.createdAt).toLocaleString()}` : "";

  return (
    <section className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-4" aria-label="API diff">
      <h2 className="text-sm font-semibold">API Diff</h2>

      <div className="flex items-center gap-2">
        <CompactSelect
          value={mode}
          onChange={(v) => {
            // SAFETY: options are exactly the two supported diff modes.
            setMode(v as DiffMode);
          }}
          options={MODE_OPTIONS}
          ariaLabel="Diff mode"
        />
      </div>

      {mode === "specs" && (
        <div className="flex flex-wrap items-end gap-2">
          <div className="flex min-w-56 flex-1 flex-col gap-1">
            <label htmlFor="diff-path-a" className="text-xs font-medium">Spec A</label>
            <Input
              id="diff-path-a"
              value={pathA}
              onChange={(e) => setPathA(e.target.value)}
              placeholder="specs/old.yaml"
              spellCheck={false}
              className="font-mono text-xs"
            />
          </div>
          <div className="flex min-w-56 flex-1 flex-col gap-1">
            <label htmlFor="diff-path-b" className="text-xs font-medium">Spec B</label>
            <Input
              id="diff-path-b"
              value={pathB}
              onChange={(e) => setPathB(e.target.value)}
              placeholder="specs/new.yaml"
              spellCheck={false}
              className="font-mono text-xs"
            />
          </div>
          <Button size="sm" disabled={busy || pathA.trim() === "" || pathB.trim() === ""} onClick={() => void run()}>
            {busy ? <Spinner data-icon="inline-start" /> : <GitCompareArrows data-icon="inline-start" />}
            Diff
          </Button>
        </div>
      )}

      {mode === "responses" && (
        <div className="flex flex-wrap items-end gap-2">
          <div className="flex min-w-64 flex-1 flex-col gap-1">
            <label htmlFor="diff-entry-a" className="text-xs font-medium">Entry A</label>
            <select
              id="diff-entry-a"
              value={entryA?.id ?? ""}
              onChange={(e) => setEntryA(entries.find((x) => x.id === e.target.value) ?? null)}
              className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
            >
              <option value="">Pick entry…</option>
              {entries.map((e) => (
                <option key={e.id} value={e.id}>{entryOption(e)}</option>
              ))}
            </select>
          </div>
          <div className="flex min-w-64 flex-1 flex-col gap-1">
            <label htmlFor="diff-entry-b" className="text-xs font-medium">Entry B</label>
            <select
              id="diff-entry-b"
              value={entryB?.id ?? ""}
              onChange={(e) => setEntryB(entries.find((x) => x.id === e.target.value) ?? null)}
              className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
            >
              <option value="">Pick entry…</option>
              {entries.map((e) => (
                <option key={e.id} value={e.id}>{entryOption(e)}</option>
              ))}
            </select>
          </div>
          <Button size="sm" disabled={busy || !entryA || !entryB || entryA.id === entryB?.id} onClick={() => void run()}>
            {busy ? <Spinner data-icon="inline-start" /> : <GitCompareArrows data-icon="inline-start" />}
            Diff
          </Button>
        </div>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {specResult && mode === "specs" && specResult.result && (
        <div className="flex min-h-0 flex-col gap-2">
          <div className="flex items-center gap-2 text-xs">
            {specResult.breaking > 0 && (
              <Badge variant="outline" className="border-status-error/40 text-status-error">
                {specResult.breaking} breaking
              </Badge>
            )}
            {specResult.addition > 0 && (
              <Badge variant="outline" className="border-status-ok/40 text-status-ok">
                {specResult.addition} additions
              </Badge>
            )}
            {specResult.result.hasChanges && specResult.breaking === 0 && (
              <span className="text-muted-foreground">no breaking changes detected</span>
            )}
          </div>
          <ChangesList result={specResult.result} severity />
        </div>
      )}

      {respResult && mode === "responses" && respResult.metaA && respResult.metaB && (
        <div className="flex min-h-0 flex-col gap-2">
          <div className="grid grid-cols-2 gap-2 text-[11px]">
            {[respResult.metaA, respResult.metaB].map((m, i) => (
              <div key={`${m.id}-${i}`} className="rounded-md border border-border px-2 py-1.5 font-mono">
                <p className={m.status >= 400 ? "text-status-error" : "text-status-ok"}>
                  {m.method} {m.status}
                </p>
                <p className="truncate text-muted-foreground">{m.url}</p>
                {m.env !== "" && <p className="text-muted-foreground">env: {m.env}</p>}
              </div>
            ))}
          </div>
          {respResult.result && <ChangesList result={respResult.result} severity={false} />}
        </div>
      )}
    </section>
  );
}
