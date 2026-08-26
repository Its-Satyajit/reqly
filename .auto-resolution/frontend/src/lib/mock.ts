export interface MockRoute {
  /** Stable identity for list rendering and updates. */
  id?: string;
  method: string;
  path: string;
  status: number;
  body: string;
  /** Raw header lines ("Key: value"), parsed by the bridge. */
  headerLines: string[];
  enabled: boolean;
}

export interface MockStatus {
  running: boolean;
  url?: string;
  port?: number;
  error?: string;
}

export interface MockAdapter {
  start(input: {
    specPath?: string;
    port: number;
    delayMs?: number;
    failEvery?: number;
    routes: {
      method: string;
      path: string;
      status: number;
      body: string;
      headers?: Record<string, string>;
      enabled: boolean;
    }[];
  }): Promise<MockStatus>;
  stop(): Promise<void>;
  status(): Promise<MockStatus>;
}

/** Parses "Key: value" lines into a header map (invalid lines skipped). */
/** Mock header map as consumed by the bridge; distinct from response HeaderMap. */
export interface MockHeaders {
  [header: string]: string;
}

export const parseHeaderLines = (lines: string[]): MockHeaders => {
  const headers: MockHeaders = {};
  for (const line of lines) {
    const idx = line.indexOf(":");
    if (idx <= 0) continue;
    const key = line.slice(0, idx).trim();
    if (key === "") continue;
    headers[key] = line.slice(idx + 1).trim();
  }
  return headers;
}

/** Renders a header map back into editable lines. */
export function headerLinesFrom(headers?: MockHeaders): string[] {
  return Object.entries(headers ?? {}).map(([k, v]) => `${k}: ${v}`);
}

export const fallbackMockAdapter: MockAdapter = {
  async start() {
    throw new Error("mock server is not available in this build");
  },
  async stop() {},
  async status() {
    return { running: false };
  },
};

export const MOCK_METHOD_OPTIONS = [
  { value: "", label: "Any" },
  { value: "GET", label: "GET" },
  { value: "POST", label: "POST" },
  { value: "PUT", label: "PUT" },
  { value: "PATCH", label: "PATCH" },
  { value: "DELETE", label: "DELETE" },
];
