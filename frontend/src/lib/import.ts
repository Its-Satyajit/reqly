export type ImportFormat =
  | ""
  | "curl"
  | "openapi"
  | "har"
  | "postman"
  | "insomnia"
  | "bruno";

export interface ImportReportEntry {
  itemPath: string;
  category: string;
  severity: string;
  message: string;
}

export interface ImportReport {
  importer: string;
  entries: ImportReportEntry[] | null;
}

export interface ImportedOperation {
  method: string;
  path: string;
  operationId?: string;
  tags?: string[] | null;
  summary?: string;
}

export interface ImportedRequest {
  method?: string;
  url?: string;
  headers?: { key: string; value: string }[] | null;
  query?: { key: string; value: string }[] | null;
  body?: string;
}

export interface ImportOutcome {
  kind: string;
  format: string;
  title?: string;
  requestCount: number;
  environmentCount?: number;
  targetDir?: string;
  report?: ImportReport | null;
  operations?: ImportedOperation[] | null;
  request?: ImportedRequest | null;
}

export interface ImportAdapter {
  detect(content: string): Promise<{ format: string; ok: boolean }>;
  preview(input: { content: string; formatHint?: string }): Promise<ImportOutcome>;
  commit(input: {
    content: string;
    formatHint?: string;
    targetDir?: string;
  }): Promise<ImportOutcome>;
}

export const fallbackImportAdapter: ImportAdapter = {
  async detect() {
    return { format: "", ok: false };
  },
  async preview() {
    throw new Error("import is not available in this build");
  },
  async commit() {
    throw new Error("import is not available in this build");
  },
};

export const IMPORT_FORMAT_OPTIONS = [
  { value: "", label: "Auto-detect" },
  { value: "curl", label: "cURL" },
  { value: "openapi", label: "OpenAPI 3.x" },
  { value: "har", label: "HAR 1.2" },
  { value: "postman", label: "Postman v2.1" },
  { value: "insomnia", label: "Insomnia v4/v5" },
  { value: "bruno", label: "Bruno" },
];

const FORMAT_LABELS = new Map<string, string>([
  ["curl", "cURL"],
  ["openapi", "OpenAPI 3.x"],
  ["har", "HAR 1.2"],
  ["postman", "Postman v2.1"],
  ["insomnia", "Insomnia"],
  ["bruno", "Bruno"],
]);

export function formatLabel(format: string): string {
  return FORMAT_LABELS.get(format) ?? format;
}
