import { create } from "zustand";
import {
  fallbackMockAdapter,
  headerLinesFrom,
  parseHeaderLines,
  pruneExpiredState,
  type MockAdapter,
  type MockRoute,
  type MockStatus,
  type MockScenario,
  type MockStateVariable,
  type FaultInjection,
  type RequestMatcher,
  type MockLogEntry,
} from "../lib/mock";

let routeSeq = 0;
let scenarioSeq = 0;
let matcherSeq = 0;

/** nextRouteId gives manual routes a stable React key. */
const nextRouteId = () => `route-${++routeSeq}`;
const nextScenarioId = () => `scenario-${++scenarioSeq}`;
const nextMatcherId = () => `matcher-${++matcherSeq}`;

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
  // §56.7 — Scenarios
  scenarios: MockScenario[]
  activeScenarioId: string | null
  // §56.7 — State variables
  stateVariables: MockStateVariable[]
  // §56.7 — Fault injection
  faultInjection: FaultInjection
  // §56.7 — Request matching
  requestMatchers: RequestMatcher[]
  // §56.7 — Logs
  logs: MockLogEntry[]
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
  // §56.7 — Scenario actions
  setActiveScenario(id: string | null): void
  createScenario(name: string): void
  deleteScenario(id: string): void
  updateScenario(id: string, patch: Partial<MockScenario>): void
  // §56.7 — State actions
  setMockStateVariable(key: string, value: string, ttl?: number): void
  clearMockStateVariable(key: string): void
  pruneExpiredState(): void
  // §56.7 — Fault injection
  setFaultInjection(patch: Partial<FaultInjection>): void
  // §56.7 — Request matching
  addRequestMatcher(matcher: Omit<RequestMatcher, "id">): void
  removeRequestMatcher(id: string): void
  updateRequestMatcher(id: string, patch: Partial<RequestMatcher>): void
  // §56.7 — Logs
  addMockLog(entry: MockLogEntry): void
  clearMockLogs(): void
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
  scenarios: [],
  activeScenarioId: null,
  stateVariables: [],
  faultInjection: { enabled: false, type: "delay", probability: 0 },
  requestMatchers: [],
  logs: [],

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

  // §56.7 — Scenario actions
  setActiveScenario(id) {
    set({ activeScenarioId: id });
  },

  createScenario(name) {
    const id = nextScenarioId();
    const scenario: MockScenario = { id, name, routes: [], variables: {} };
    set((s) => ({ scenarios: [...s.scenarios, scenario] }));
  },

  deleteScenario(id) {
    set((s) => ({
      scenarios: s.scenarios.filter((sc) => sc.id !== id),
      activeScenarioId: s.activeScenarioId === id ? null : s.activeScenarioId,
    }));
  },

  updateScenario(id, patch) {
    set((s) => ({
      scenarios: s.scenarios.map((sc) => (sc.id === id ? { ...sc, ...patch } : sc)),
    }));
  },

  // §56.7 — State actions
  setMockStateVariable(key, value, ttl) {
    set((s) => {
      const existing = s.stateVariables.findIndex((v) => v.key === key);
      const updated: MockStateVariable = { key, value, ttl, updatedAt: Date.now() };
      if (existing >= 0) {
        const copy = [...s.stateVariables];
        copy[existing] = updated;
        return { stateVariables: copy };
      }
      return { stateVariables: [...s.stateVariables, updated] };
    });
  },

  clearMockStateVariable(key) {
    set((s) => ({
      stateVariables: s.stateVariables.filter((v) => v.key !== key),
    }));
  },

  pruneExpiredState() {
    set((s) => ({ stateVariables: pruneExpiredState(s.stateVariables) }));
  },

  // §56.7 — Fault injection
  setFaultInjection(patch) {
    set((s) => ({ faultInjection: { ...s.faultInjection, ...patch } }));
  },

  // §56.7 — Request matching
  addRequestMatcher(matcher) {
    const id = nextMatcherId();
    set((s) => ({
      requestMatchers: [...s.requestMatchers, { ...matcher, id }],
    }));
  },

  removeRequestMatcher(id) {
    set((s) => ({
      requestMatchers: s.requestMatchers.filter((m) => m.id !== id),
    }));
  },

  updateRequestMatcher(id, patch) {
    set((s) => ({
      requestMatchers: s.requestMatchers.map((m) => (m.id === id ? { ...m, ...patch } : m)),
    }));
  },

  // §56.7 — Logs
  addMockLog(entry) {
    set((s) => ({ logs: [...s.logs, entry] }));
  },

  clearMockLogs() {
    set({ logs: [] });
  },
}));

// Re-exported for tests and future callers that need line rendering.
export { headerLinesFrom };
