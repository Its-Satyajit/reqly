export interface RunnerStepView {
  index: number;
  status?: number;
  error?: string;
  url?: string;
  bodyPreview?: string;
}

export interface PaginationConfigInput {
  strategy: "page" | "offset" | "cursor" | "link-header";
  pageParam?: string;
  pageSizeParam?: string;
  offsetParam?: string;
  limitParam?: string;
  cursorParam?: string;
  nextPath?: string;
  maxPages?: number;
  pageSize?: number;
  limit?: number;
}

/** RunnerSummary is the `.done` event payload (string-keyed scalars from Go). */
export interface RunnerSummary {
  [field: string]: string | number | boolean;
}

export interface RunnerAdapter {
  start(input: {
    runId: string;
    kind: "pagination" | "bulk";
    request: unknown; // TabDraft JSON
    pagination?: PaginationConfigInput;
    maxPagesOverride?: number;
    data?: string;
    dataFormat?: string;
    parallel?: boolean;
    concurrency?: number;
  }): Promise<void>;
  cancel(runId: string): Promise<void>;
  subscribe(
    runId: string,
    handlers: {
      onStep: (step: RunnerStepView) => void;
      onDone: (summary: RunnerSummary) => void;
    },
  ): () => void;
}

// Bridge registry, same pattern as the other feature adapters.
let bridge: RunnerAdapter | null = null;

/** setRunnerBridge installs the host adapter (called once from the bridge). */
export function setRunnerBridge(adapter: RunnerAdapter): void {
  bridge = adapter;
}

export function getRunnerBridge(): RunnerAdapter {
  if (!bridge) throw new Error("runners are not available in this build");
  return bridge;
}

let runCounter = 0;

/** nextRunId mints a fresh run id for event routing. */
export function nextRunId(prefix: string): string {
  return `${prefix}-${Date.now()}-${++runCounter}`;
}

export const STRATEGY_OPTIONS = [
  { value: "page", label: "Page (?page=N)" },
  { value: "offset", label: "Offset (?offset=N&limit=M)" },
  { value: "cursor", label: "Cursor (next via JSONPath)" },
  { value: "link-header", label: "Link header (rel=next)" },
];
