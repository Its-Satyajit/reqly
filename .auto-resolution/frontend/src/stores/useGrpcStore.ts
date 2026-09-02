import { create } from "zustand";
import {
  fallbackGrpcAdapter,
  type GrpcAdapter,
  type GrpcRequest,
  type GrpcResultView,
  type GrpcService,
  type GrpcStreamMessage,
} from "../lib/grpcclient";

const MESSAGE_BUFFER_CAP = 500;

export interface GrpcTabState {
  sessionId?: string;
  target: string;
  protoFiles: string;
  tls: boolean;
  tlsSkipVerify: boolean;
  services: GrpcService[];
  service: string;
  method: string;
  message: string;
  unaryResult: GrpcResultView | null;
  streamMessages: GrpcStreamMessage[];
  status: "idle" | "connecting" | "unary-ok" | "streaming" | "done" | "error";
  error: string | null;
}

interface GrpcState {
  adapter: GrpcAdapter
  tabs: Record<string, GrpcTabState>
  setAdapter(adapter: GrpcAdapter): void
  newTab(tabId: string): void
  update(tabId: string, patch: Partial<GrpcTabState>): void
  discover(tabId: string): Promise<void>
  send(tabId: string): Promise<void>
  stop(tabId: string): Promise<void>
}

let sessionCounter = 0;

function emptyTab(): GrpcTabState {
  return {
    target: "",
    protoFiles: "",
    tls: false,
    tlsSkipVerify: false,
    services: [],
    service: "",
    method: "",
    message: "{}",
    unaryResult: null,
    streamMessages: [],
    status: "idle",
    error: null,
  };
}

function toRequest(tab: GrpcTabState): GrpcRequest {
  let message: unknown;
  try {
    message = JSON.parse(tab.message);
  } catch {
    message = tab.message; // backend surfaces a validation error
  }
  return {
    url: tab.target.trim(),
    headers: [],
    grpc: {
      service: tab.service,
      method: tab.method,
      message,
      protoFiles: tab.protoFiles
        .split(",")
        .map((p) => p.trim())
        .filter(Boolean),
      tls: tab.tls || tab.tlsSkipVerify,
      tlsSkipVerify: tab.tlsSkipVerify,
    },
  };
}

export const useGrpcStore = create<GrpcState>((set, get) => ({
  adapter: fallbackGrpcAdapter,
  tabs: {},

  setAdapter(adapter) {
    set({ adapter });
  },

  newTab(tabId: string) {
    set((s) => ({ tabs: { ...s.tabs, [tabId]: emptyTab() } }));
  },

  update(tabId, patch) {
    set((s) => {
      const tab = s.tabs[tabId];
      if (!tab) return s;
      return { tabs: { ...s.tabs, [tabId]: { ...tab, ...patch } } };
    });
  },

  async discover(tabId) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    if (!tab || tab.target.trim() === "") return;
    get().update(tabId, { status: "connecting", error: null });
    try {
      const services = await adapter.services({
        target: tab.target.trim(),
        protoFiles: toRequest(tab).grpc.protoFiles,
      });
      get().update(tabId, { services, status: "idle", error: null });
    } catch (err) {
      get().update(tabId, {
        status: "error",
        error: err instanceof Error ? err.message : String(err),
      });
    }
  },

  async send(tabId) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    if (!tab) return;
    const selected = tab.services.find((svc) => svc.name === tab.service);
    const streaming = selected?.methods.find((m) => m.name === tab.method)?.serverStreaming ?? false;
    const sessionId = `grpc-${Date.now()}-${sessionCounter++}`;
    get().update(tabId, {
      sessionId,
      status: streaming ? "streaming" : "connecting",
      error: null,
      unaryResult: null,
      streamMessages: [],
    });

    if (!streaming) {
      try {
        const result = await adapter.invoke(sessionId, toRequest(tab));
        get().update(tabId, {
          unaryResult: result,
          status: result.ok ? "unary-ok" : "error",
          error: result.ok ? null : `gRPC status ${result.codeName ?? ""}: ${result.statusMessage ?? ""}`,
        });
      } catch (err) {
        get().update(tabId, {
          status: "error",
          error: err instanceof Error ? err.message : String(err),
        });
      }
      return;
    }

    adapter.subscribe(sessionId, (event) => {
      if (event.type === "message") {
        set((s) => {
          const cur = s.tabs[tabId];
          if (!cur) return s;
          const msgs = [...cur.streamMessages, { seq: event.seq ?? cur.streamMessages.length + 1, messageJson: event.data ?? "" }];
          if (msgs.length > MESSAGE_BUFFER_CAP) msgs.splice(0, msgs.length - MESSAGE_BUFFER_CAP);
          return { tabs: { ...s.tabs, [tabId]: { ...cur, streamMessages: msgs } } };
        });
        return;
      }
      if (event.type === "done") {
        get().update(tabId, { status: event.codeName ? "error" : "done" });
        if (event.codeName) {
          get().update(tabId, { error: `gRPC status ${event.codeName}: ${event.data ?? ""}` });
        }
        return;
      }
      if (event.type === "cancelled") {
        get().update(tabId, { status: "done" });
        return;
      }
      if (event.type === "error") {
        get().update(tabId, { status: "error", error: event.data ?? "stream error" });
      }
    });
    try {
      await adapter.stream(sessionId, toRequest(tab));
    } catch (err) {
      get().update(tabId, {
        status: "error",
        error: err instanceof Error ? err.message : String(err),
      });
    }
  },

  async stop(tabId) {
    const { adapter, tabs } = get();
    const sessionId = tabs[tabId]?.sessionId;
    if (sessionId) await adapter.cancel(sessionId);
  },
}));
