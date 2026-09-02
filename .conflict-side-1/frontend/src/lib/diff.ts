export interface DiffChange {
  type: "create" | "update" | "delete";
  path: string[];
  from?: unknown;
  to?: unknown;
  /** "breaking" | "addition" | "info" — set for spec diffs. */
  severity?: string;
}

export interface DiffResultView {
  hasChanges: boolean;
  changes: DiffChange[] | null;
}

export interface SpecDiffResult {
  result: DiffResultView;
  breaking: number;
  addition: number;
}

export interface ResponseDiffMeta {
  id: string;
  url: string;
  method: string;
  status: number;
  env?: string;
  preview: string;
}

export interface ResponseDiffResult {
  metaA: ResponseDiffMeta | null;
  metaB: ResponseDiffMeta | null;
  result: DiffResultView;
}

export interface HistoryEntryRef {
  id: string;
  url: string;
  method: string;
  status: number;
  env?: string;
  createdAt?: string;
}

export interface DiffAdapter {
  specs(pathA: string, pathB: string): Promise<SpecDiffResult>;
  responses(idA: string, idB: string): Promise<ResponseDiffResult>;
}

// Bridge registry: the desktop host installs its adapter at startup; other
// surfaces keep the throwing fallback.
let diffBridge: DiffAdapter | null = null;

/** setDiffBridge installs the host adapter (called once from the bridge). */
export function setDiffBridge(adapter: DiffAdapter): void {
  diffBridge = adapter;
}

/** getDiffBridge returns the installed adapter or a throwing fallback. */
export function getDiffBridge(): DiffAdapter {
  return diffBridge ?? fallbackDiffAdapter;
}

export const fallbackDiffAdapter: DiffAdapter = {
  async specs() {
    throw new Error("diff is not available in this build");
  },
  async responses() {
    throw new Error("diff is not available in this build");
  },
};

export function changeLabel(change: DiffChange): string {
  const verb =
    change.type === "create" ? "added" : change.type === "delete" ? "removed" : "changed";
  return `${verb} ${change.path.join(".")}`;
}
