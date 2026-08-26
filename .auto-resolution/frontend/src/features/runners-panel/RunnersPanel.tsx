import { useEffect, useState } from "react";
import { Layers, Play, Square } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Switch } from "#components/ui/switch";
import { Textarea } from "#components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "#components/ui/select";

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

type Mode = "pagination" | "bulk";


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

  useEffect(() => {
    if (activeRunId == null) return;
    const unlisten = getRunnerBridge().listen(activeRunId, {
      onStep: (step) => {
        frameSeq += 1;
        const next = { ...step, seq: frameSeq };
        setUi((prev) => ({ ...prev, steps: [...prev.steps.slice(-499), next] }));
      },
      onDone: (summary) => {
        setUi((prev) => ({ ...prev, summary, running: false }));
      },
    });
    return () => {
      unlisten();
    };
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

    // The listener lives in an effect keyed on the run id so unmount and
    // restarts always detach cleanly (react-doctor/effect-needs-cleanup).
    patch({ activeRunId: runId });
    getRunnerBridge()
      .start({
        runId,
        kind: mode,
        request: draft,
        pagination,
        maxPagesOverride: mode === "pagination" ? maxPages : undefined,
        data: mode === "bulk" ? data : undefined,
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
    <div className="flex flex-col gap-3 border-t border-border pt-3" aria-label="Pagination and bulk runners">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <Layers className="mr-1 inline size-3.5" aria-hidden />
          Runners
        </span>
        <Select
          items={[
            { value: "pagination", label: "Pagination" },
            { value: "bulk", label: "Bulk" },
          ]}
          value={mode}
          onValueChange={(v) => {
            if (v === null) return;
            // SAFETY: options are exactly the two runner kinds.
            patch({ mode: v as Mode });
          }}
        >
          <SelectTrigger aria-label="Runner mode" className="h-7 w-auto gap-1 rounded-md px-2 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent className="max-h-72 min-w-(--anchor-width)">
            {[
              { value: "pagination", label: "Pagination" },
              { value: "bulk", label: "Bulk" },
            ].map((option) => (
              <SelectItem key={option.value} value={option.value} className="text-xs">
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {activeTabId == null && (
        <p className="text-[11px] text-muted-foreground">
          Open a request tab to configure a run against it.
        </p>
      )}

      {mode === "pagination" ? (
        <>
          <div className="flex items-center gap-1.5">
            <span className="text-xs font-medium">Strategy</span>
            <Select
              items={STRATEGY_OPTIONS}
              value={strategy}
              onValueChange={(v) => {
                if (v === null) return;
                // SAFETY: options are the four core pagination strategies.
                patch({ strategy: v as PaginationConfigInput["strategy"] });
              }}
            >
              <SelectTrigger aria-label="Pagination strategy" className="h-7 w-auto gap-1 rounded-md px-2 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-72 min-w-(--anchor-width)">
                {STRATEGY_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value} className="text-xs">
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
          <Textarea
            value={data}
            onChange={(e) => patch({ data: e.target.value })}
            rows={3}
            spellCheck={false}
            aria-label="Bulk data rows (CSV or JSON array)"
            placeholder={"id,name\n1,ada"}
            className="resize-y font-mono text-[11px]"
          />
          <div className="flex items-center gap-3 text-xs">
            <label className="flex items-center gap-1.5">
              <Switch
                checked={parallel}
                onCheckedChange={(checked) => patch({ parallel: checked })}
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
  );
}
