export type ExportFormat = "postman" | "openapi" | "har" | "workspace" | "environment";

export interface ExportOutcome {
  format: ExportFormat;
  path: string;
  requestCount?: number;
  entryCount?: number;
}

export interface ExportAdapter {
  run(input: {
    format: ExportFormat;
    collection?: string;
    outName?: string;
  }): Promise<ExportOutcome>;
}

export const fallbackExportAdapter: ExportAdapter = {
  async run() {
    throw new Error("export is not available in this build");
  },
};

export const EXPORT_FORMAT_OPTIONS = [
  { value: "postman", label: "Postman v2.1" },
  { value: "openapi", label: "OpenAPI 3.x" },
  { value: "har", label: "HAR 1.2 (history)" },
  { value: "workspace", label: "Workspace copy" },
  { value: "environment", label: "Environment YAML" },
];

const FORMAT_LABELS = new Map<string, string>([
  ["postman", "Postman v2.1"],
  ["openapi", "OpenAPI 3.x"],
  ["har", "HAR 1.2"],
  ["workspace", "Workspace copy"],
  ["environment", "Environment YAML"],
]);

export function exportFormatLabel(format: string): string {
  return FORMAT_LABELS.get(format) ?? format;
}
