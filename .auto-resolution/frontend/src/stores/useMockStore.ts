import { create } from "zustand";
import {
  fallbackMockAdapter,
  headerLinesFrom,
  parseHeaderLines,
  type MockAdapter,
  type MockRoute,
  type MockStatus,
} from "../lib/mock";

let routeSeq = 0;

/** nextRouteId gives manual routes a stable React key. */
const nextRouteId = () => `route-${++routeSeq}`;

interface MockState {
  adapter: MockAdapter
  specPath: string
  port: number
  delayMs: number
  failEvery: number
  routes: MockRoute[]
  status: MockStatus
  busy: boolean
  error: string | null
  setAdapter(adapter: MockAdapter): void
  setSpecPath(specPath: string): void
  setPort(port: number): void
  setDelayMs(delayMs: number): void
  setFailEvery(failEvery: number): void
  updateRoute(index: number, patch: Partial<MockRoute>): void
  addRoute(): void
  removeRoute(index: number): void
  start(): Promise<void>
  stop(): Promise<void>
  refreshStatus(): Promise<void>
}

const emptyStatus: MockStatus = { running: false };

function routeToPayload(r: MockRoute) {
  return {
    method: r.method,
    path: r.path,
    status: r.status || 200,
    body: r.body,
    headers: parseHeaderLines(r.headerLines),
    enabled: r.enabled,
  };
}

export const useMockStore = create<MockState>((set) => ({
  adapter: fallbackMockAdapter,
  specPath: "",
  port: 4010,
  delayMs: 0,
  failEvery: 0,
  routes: [
    {
      id: "route-seed",
      method: "GET",
      path: "/hello",
      status: 200,
      body: '{"message":"hello"}',
      headerLines: [],
      enabled: true,
    },
  ],
  status: emptyStatus,
  busy: false,
  error: null,

  setAdapter(adapter) {
    set({ adapter });
    void useMockStore.getState().refreshStatus();
  },

  setSpecPath(specPath) {
    set({ specPath });
  },

  setPort(port) {
    set({ port });
  },

  setDelayMs(delayMs) {
    set({ delayMs });
  },

  setFailEvery(failEvery) {
    set({ failEvery });
  },

  updateRoute(index, patch) {
    set((s) => ({
      routes: s.routes.map((r, i) => (i === index ? { ...r, ...patch } : r)),
    }));
  },

  addRoute() {
    set((s) => ({
      routes: [
        ...s.routes,
        {
          id: nextRouteId(),
          method: "GET",
          path: "/new",
          status: 200,
          body: "{}",
          headerLines: [],
          enabled: true,
        },
      ],
    }));
  },

  removeRoute(index) {
    set((s) => ({ routes: s.routes.filter((_, i) => i !== index) }));
  },

  async start() {
    const s = useMockStore.getState();
    set({ busy: true, error: null });
    try {
      const status = await s.adapter.start({
        specPath: s.specPath.trim() === "" ? undefined : s.specPath.trim(),
        port: s.port,
        delayMs: s.delayMs > 0 ? s.delayMs : undefined,
        failEvery: s.failEvery > 1 ? s.failEvery : undefined,
        routes: s.routes.map(routeToPayload),
      });
      set({ status, busy: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : String(err),
        busy: false,
      });
    }
  },

  async stop() {
    set({ busy: true, error: null });
    try {
      await useMockStore.getState().adapter.stop();
      set({ status: emptyStatus, busy: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : String(err),
        busy: false,
      });
    }
  },

  async refreshStatus() {
    try {
      const status = await useMockStore.getState().adapter.status();
      set({ status });
    } catch {
      /* keep last known state */
    }
  },
}));

// Re-exported for tests and future callers that need line rendering.
export { headerLinesFrom };
