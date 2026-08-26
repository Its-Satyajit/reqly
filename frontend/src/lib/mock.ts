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

// --- §56.7 Extended Mock Server GUI types ---

export interface MockScenario {
  id: string;
  name: string;
  description?: string;
  routes: MockRoute[];
  variables?: { [key: string]: string };
}

export interface MockStateVariable {
  key: string;
  value: string;
  /** TTL in ms. 0 = permanent. Undefined = permanent. */
  ttl?: number;
  updatedAt: number;
}

export interface FaultInjection {
  enabled: boolean;
  type: "delay" | "drop" | "error" | "corrupt";
  probability: number;
  delayMs?: number;
  errorCode?: number;
  errorMessage?: string;
}

export interface RequestMatcher {
  id: string;
  method?: string;
  pathPattern: string;
  headers?: MockHeaders;
  bodyPattern?: string;
  priority: number;
}

export interface MockLogEntry {
  timestamp: number;
  method: string;
  path: string;
  status: number;
  duration: number;
  matchedRoute?: string;
  scenario?: string;
  error?: string;
}

let scenarioSeq = 0;

/** Create a new mock scenario with a stable id. */
export function createMockScenario(name: string, description?: string): MockScenario {
  return {
    id: `scenario-${++scenarioSeq}`,
    name,
    description,
    routes: [],
    variables: {},
  };
}

/** Match a request against a list of matchers; returns highest-priority (lowest number) match. */
export function matchRoute(
  method: string,
  path: string,
  matchers: RequestMatcher[],
): RequestMatcher | null {
  const sorted = [...matchers].sort((a, b) => a.priority - b.priority);
  for (const m of sorted) {
    if (m.method && m.method.toUpperCase() !== method.toUpperCase()) continue;
    // Simple glob: exact match or prefix with **
    if (m.pathPattern.endsWith("/**")) {
      const prefix = m.pathPattern.slice(0, -3);
      if (path.startsWith(prefix)) return m;
    } else if (m.pathPattern === path) {
      return m;
    }
  }
  return null;
}

/** Remove state variables whose TTL has expired. */
export function pruneExpiredState(vars: MockStateVariable[]): MockStateVariable[] {
  const now = Date.now();
  return vars.filter((v) => {
    if (v.ttl === undefined || v.ttl === 0) return true;
    return now - v.updatedAt < v.ttl;
  });
}
