import { useEffect, useState } from "react";
import { Layers, Play, Square } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { CompactSelect } from "#components/CompactSelect";

/** inputInt parses numeric input text, falling back on non-numeric values. */
function inputInt(value: string, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}
import {
  getRunnerBridge,
  nextRunId,
  STRATEGY_OPTIONS,
  type RunnerSummary,
  type PaginationConfigInput,
  type RunnerStepView,
} from "#lib/runners";
import { useRequestStore, type TabDraft } from "#stores/useRequestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useDatasetStore } from "#stores/useDatasetStore";
import { DatasetPicker } from "./DatasetPicker";

type Mode = "pagination" | "bulk" | "graph";


function draftOf(activeTabId: string | null): TabDraft | null {
  if (!activeTabId) return null;
  return useRequestStore.getState().drafts[activeTabId] ?? null;
}

let frameSeq = 0;

export function RunnersPanel() {
  const activeTabId = useWorkspaceStore((s) => s.activeTabId);
  // One state object keeps the panel's related fields together
  // (react-doctor/prefer-useReducer). steps/summary/running/error update via
  // setUi() from the runner event handlers.
  // SAFETY: each `as` below only pins a string literal to its declared
  // union member; the object shape is otherwise fully inferred.
  const [ui, setUi] = useState({
    mode: "pagination" as Mode,
    strategy: "page" as PaginationConfigInput["strategy"],
    maxPages: 20,
    nextPath: "$.nextCursor",
    data: "id\n1\n2\n3",
    parallel: false,
    concurrency: 4,
    steps: [] as RunnerStepView[],
    summary: null as RunnerSummary | null,
    running: false,
    error: null as string | null,
    activeRunId: null as string | null,
  });
  const patch = (p: Partial<typeof ui>) => setUi((prev) => ({ ...prev, ...p }));
  const {
    mode,
    strategy,
    maxPages,
    nextPath,
    data,
    parallel,
    concurrency,
    steps,
    summary,
    running,
    error,
    activeRunId,
  } = ui;

  // Track active run listener
  useEffect(() => {
    if (!activeRunId) return;
    const unsub = getRunnerBridge().listen(activeRunId, {
      onStep: (step) => {
        frameSeq += 1;
        const next = { ...step, seq: frameSeq };
        setUi((prev) => ({ ...prev, steps: [...prev.steps.slice(-499), next] }));
      },
      onDone: (summary) => {
        setUi((prev) => ({ ...prev, summary, running: false }));
      },
    });
    return () => unsub();
  }, [activeRunId]);

  // Draft is captured at start; the panel shows live progress per step.
  const start = (): void => {
    const draft = draftOf(activeTabId);
    if (!draft) {
      patch({ error: "Open a request tab first — runners execute the active request." });
      return;
    }
    const runId = nextRunId(mode);
    patch({ steps: [], summary: null, error: null, running: true });

    let pagination: PaginationConfigInput | undefined;
    if (mode === "pagination") {
      pagination = { strategy, maxPages };
      if (strategy === "cursor") pagination.nextPath = nextPath;
    }

    // Prefer dataset store (picker) over legacy textarea — handles CSV/JSON validation & preview.
    const bulkData = (() => {
      if (mode !== "bulk") return undefined;
      const ds = useDatasetStore.getState().dataset;
      if (ds?.rawContent && ds.rawContent.trim() !== "") return ds.rawContent;
      return data.trim() !== "" ? data : undefined;
    })();

    // The listener lives in an effect keyed on the run id so unmount and
    // restarts always detach cleanly (react-doctor/effect-needs-cleanup).
    patch({ activeRunId: runId });
    getRunnerBridge()
      .start({
        runId,
        kind: mode === "bulk" ? "bulk" : "pagination",
        request: draft,
        pagination,
        maxPagesOverride: mode === "pagination" ? maxPages : undefined,
        data: bulkData,
        parallel: mode === "bulk" ? parallel : undefined,
        concurrency: mode === "bulk" && parallel ? concurrency : undefined,
      })
      .catch((e) => {
        patch({
          error: e instanceof Error ? e.message : String(e),
          running: false,
        });
      });
  };

  const stop = (): void => {
    if (activeRunId != null) void getRunnerBridge().cancel(activeRunId);
    patch({ running: false });
  };

  return (
    <div className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="Pagination and bulk runners">
      <PageHeader
        icon={Layers}
        title="Runners & Execution Graph"
        description="Paginate through multi-page endpoints, bulk-execute requests, or visualize request dependency chains"
        actions={
          <CompactSelect
            value={mode}
            onChange={(v) => {
              // SAFETY: options are exactly the three runner modes.
              patch({ mode: v as Mode });
            }}
            options={[
              { value: "pagination", label: "Pagination" },
              { value: "bulk", label: "Bulk" },
              { value: "graph", label: "Dependency Graph" },
            ]}
          ariaLabel="Runner mode"
          />
        }
      />
      <div className="flex flex-col gap-3 p-3">
      {mode === "graph" ? (
        <div className="rounded-lg border border-border bg-card p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-semibold">Workspace Request Dependency Graph (§57.4)</h3>
            <span className="text-[11px] font-mono text-muted-foreground">Variable propagation & chained flow</span>
          </div>
          <div className="rounded border border-border/50 bg-background/60 p-4 font-mono text-xs space-y-4">
            <div className="flex items-center gap-2">
              <span className="rounded bg-primary/10 text-primary px-2 py-1 font-semibold">1. Auth / Token Exchange</span>
              <span className="text-muted-foreground">──────►</span>
              <span className="text-xs text-status-info">sets &#123;&#123;token&#125;&#125;</span>
            </div>
            <div className="flex items-center gap-2 pl-6">
              <span className="text-muted-foreground">│</span>
            </div>
            <div className="flex items-center gap-2 pl-6">
              <span className="text-muted-foreground">├──────►</span>
              <span className="rounded bg-muted px-2 py-1 font-medium">2. Users API (GET /users)</span>
              <span className="text-muted-foreground">──────►</span>
              <span className="text-xs text-status-info">extracts &#123;&#123;userId&#125;&#125;</span>
            </div>
            <div className="flex items-center gap-2 pl-6">
              <span className="text-muted-foreground">│</span>
            </div>
            <div className="flex items-center gap-2 pl-6">
              <span className="text-muted-foreground">└──────►</span>
              <span className="rounded bg-muted px-2 py-1 font-medium">3. Orders API (POST /orders)</span>
              <span className="text-muted-foreground">──────►</span>
              <span className="text-xs text-status-info">extracts &#123;&#123;orderId&#125;&#125;</span>
            </div>
            <div className="flex items-center gap-2 pl-12">
              <span className="text-muted-foreground">└──────►</span>
              <span className="rounded bg-muted px-2 py-1 font-medium">4. Payments API (POST /payments)</span>
            </div>
          </div>
          <p className="text-xs text-muted-foreground">
            Chained variables from tests & scripts are automatically mapped into subsequent requests.
          </p>
        </div>
      ) : (
        <>
          {activeTabId == null && (
            <p className="text-[11px] text-muted-foreground">
              Open a request tab to configure a run against it.
            </p>
          )}

          {mode === "pagination" ? (
            <>
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium">Strategy</span>
                <CompactSelect
                  value={strategy}
                  onChange={(v) => {
                    // SAFETY: options are the four core pagination strategies.
                    patch({ strategy: v as PaginationConfigInput["strategy"] });
                  }}
                  options={STRATEGY_OPTIONS}
                  ariaLabel="Pagination strategy"
                />
              </div>
              {strategy === "cursor" && (
                <div className="flex items-center gap-1.5">
                  <label htmlFor="runner-nextpath" className="shrink-0 text-xs font-medium">
                    nextPath
                  </label>
                  <Input
                    id="runner-nextpath"
                    value={nextPath}
                    onChange={(e) => patch({ nextPath: e.target.value })}
                    spellCheck={false}
                    className="font-mono text-xs"
                  />
                </div>
              )}
              <div className="flex items-center gap-1.5">
                <label htmlFor="runner-maxpages" className="shrink-0 text-xs font-medium">
                  Max pages
                </label>
                <input
                  id="runner-maxpages"
                  type="number"
                  min={1}
                  value={maxPages}
                  onChange={(e) => patch({ maxPages: inputInt(e.target.value, maxPages) })}
                  className="w-24 rounded-md border border-border bg-transparent px-2 py-1 text-xs font-mono"
                />
                <span className="text-[11px] text-muted-foreground">stop condition + empty-body guard</span>
              </div>
            </>
          ) : (
            <>
              <DatasetPicker />
              {/* Legacy textarea kept for quick paste — synced with dataset store */}
              <div className="flex flex-col gap-1.5">
                <span className="text-xs font-medium text-muted-foreground">Quick paste (syncs with picker above)</span>
                <Textarea
                  value={data}
                  onChange={(e) => {
                    patch({ data: e.target.value });
                    // Keep dataset store in sync for validation/preview
                    if (e.target.value.trim() !== "") {
                      useDatasetStore.getState().loadDataset(e.target.value, "paste.csv");
                    } else {
                      useDatasetStore.getState().clearDataset();
                    }
                  }}
                  rows={3}
                  spellCheck={false}
                  aria-label="Bulk data rows (CSV or JSON array)"
                  placeholder={"id,name\n1,ada\n2,grace"}
                  className="resize-y font-mono text-[11px]"
                />
              </div>
              <div className="flex items-center gap-3 text-xs">
                <label className="flex items-center gap-1.5">
                  <input
                    type="checkbox"
                    checked={parallel}
                    onChange={(e) => patch({ parallel: e.target.checked })}
                  />
                  Parallel
                </label>
                {parallel && (
                  <label className="flex items-center gap-1">
                    Concurrency
                    <input
                      id="runner-concurrency"
                      type="number"
                      min={1}
                      value={concurrency}
                      onChange={(e) => patch({ concurrency: inputInt(e.target.value, concurrency) })}
                      className="w-16 rounded-md border border-border bg-transparent px-1 py-0.5 font-mono"
                    />
                  </label>
                )}
              </div>
            </>
          )}

          <div className="flex items-center gap-1.5">
            {running ? (
              <Button variant="outline" size="sm" onClick={stop}>
                <Square data-icon="inline-start" />
                Stop
              </Button>
            ) : (
              <Button size="sm" disabled={activeTabId == null} onClick={() => start()}>
                {running ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                Run
              </Button>
            )}
            {steps.length > 0 && (
              <Badge variant="secondary">{steps.length} steps</Badge>
            )}
          </div>
        </>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {summary != null && (
        <div className="rounded-md border border-border px-2 py-1.5 text-xs">
          {Object.entries(summary).flatMap(([k, v]) =>
            k === "lastBody"
              ? []
              : [
                  <span key={k} className="mr-2">
                    <span className="text-muted-foreground">{k}:</span>{" "}
                    <span className="font-mono">{String(v)}</span>
                  </span>,
                ],
          )}
        </div>
      )}

      {steps.length > 0 && (
        <ul className="flex flex-col gap-0.5">
          {steps.map((st) => (
            <li key={st.seq} className="font-mono text-[11px]">
              <span className="mr-1.5 text-muted-foreground">#{st.index}</span>
              {st.error ? (
                <span className="text-status-error">{st.error}</span>
              ) : (
                <>
                  <span className={st.status != null && st.status >= 400 ? "text-status-error" : "text-status-ok"}>
                    {st.status ?? "?"}
                  </span>
                  {st.url != null && st.url !== "" && (
                    <span className="ml-1 truncate text-muted-foreground">{st.url}</span>
                  )}
                </>
              )}
            </li>
          ))}
        </ul>
      )}
      </div>
    </div>
  );
}
