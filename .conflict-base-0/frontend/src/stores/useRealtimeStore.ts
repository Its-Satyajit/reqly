import { create } from "zustand";
import {
  fallbackRealtimeAdapter,
  MESSAGE_BUFFER_CAP,
  type RealtimeAdapter,
  type RealtimeFrameView,
  type RealtimeKind,
  type RealtimeStatus,
} from "../lib/realtime";

export interface RealtimeTabState {
  /** Backend session id for the live connection, set on successful open. */
  sessionId?: string;
  kind: RealtimeKind;
  url: string;
  headers: { key: string; value: string; enabled: boolean }[];
  status: RealtimeStatus;
  frames: RealtimeFrameView[];
  error: string | null;
}

interface RealtimeState {
  adapter: RealtimeAdapter
  tabs: Record<string, RealtimeTabState>
  setAdapter(adapter: RealtimeAdapter): void
  update(tabId: string, patch: Partial<RealtimeTabState>): void
  newTab(tabId: string, kind: RealtimeKind): void
  connect(tabId: string): Promise<void>
  send(tabId: string, data: string): Promise<void>
  sendBinary(tabId: string, base64: string): Promise<void>
  disconnect(tabId: string): Promise<void>
  closeTab(tabId: string): void
  appendFrame(tabId: string, frame: RealtimeFrameView): void
}

function emptyTab(kind: RealtimeKind = "ws"): RealtimeTabState {
  return { kind, url: "", headers: [], status: "idle", frames: [], error: null };
}

let sessionCounter = 0;
let frameSeq = 0;

export const useRealtimeStore = create<RealtimeState>((set, get) => ({
  adapter: fallbackRealtimeAdapter,
  tabs: {},

  setAdapter(adapter) {
    set({ adapter });
  },

  newTab(tabId, kind) {
    set((s) => ({ tabs: { ...s.tabs, [tabId]: emptyTab(kind) } }));
  },

  update(tabId, patch) {
    set((s) => {
      const tab = s.tabs[tabId];
      if (!tab) return s;
      return { tabs: { ...s.tabs, [tabId]: { ...tab, ...patch } } };
    });
  },

  async connect(tabId) {
    const { adapter, tabs } = get();
    let tab = tabs[tabId];
    if (!tab) {
      tab = emptyTab();
      set((s) => ({ tabs: { ...s.tabs, [tabId]: tab } }));
    }
    if (tab.url.trim() === "") return;
    // Fresh session id per connection so stale subscriptions never leak.
    const sessionId = `${tab.kind}-${Date.now()}-${sessionCounter++}`;
    get().update(tabId, { status: "connecting", error: null });
    adapter.subscribe(tabId, (frame) => {
      if (frame.type === "status" && frame.data === "connected") {
        get().update(tabId, { status: "connected" });
      }
      if (frame.type === "closed") {
        get().update(tabId, { status: "closed" });
      }
      if (frame.type === "error") {
        get().update(tabId, { status: "error", error: frame.data ?? "stream error" });
      }
      if (frame.type === "message") {
        get().appendFrame(tabId, frame);
      }
    });
    try {
      await adapter.open({
        sessionId,
        kind: tab.kind,
        url: tab.url.trim(),
        headers: tab.headers.filter((h) => h.enabled && h.key.trim() !== ""),
      });
      // The backend streams over the sessionId channel; remember it so send()
      // targets the live connection.
      set((s) => {
        const cur = s.tabs[tabId];
        if (!cur) return s;
        return { tabs: { ...s.tabs, [tabId]: { ...cur, sessionId } } };
      });
    } catch (err) {
      get().update(tabId, {
        status: "error",
        error: err instanceof Error ? err.message : String(err),
      });
    }
  },

  async send(tabId, data) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    const sessionId = tab?.sessionId;
    if (!sessionId) return;
    try {
      await adapter.send(sessionId, data);
    } catch (err) {
      get().update(tabId, {
        error: err instanceof Error ? err.message : String(err),
      });
    }
  },

  async sendBinary(tabId, base64) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    const sessionId = tab?.sessionId;
    if (!sessionId) return;
    if (!adapter.sendBinary) {
      get().update(tabId, { error: "binary send is not supported by this adapter" });
      return;
    }
    try {
      await adapter.sendBinary(sessionId, base64);
    } catch (err) {
      get().update(tabId, {
        error: err instanceof Error ? err.message : String(err),
      });
    }
  },

  async disconnect(tabId) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    if (!tab?.sessionId) return;
    try {
      await adapter.close(tab.sessionId);
    } catch {
      /* the closed event will settle the status */
    }
  },

  closeTab(tabId) {
    void get().disconnect(tabId);
    set((s) => {
      const tabs = { ...s.tabs };
      delete tabs[tabId];
      return { tabs };
    });
  },

  appendFrame(tabId, frame) {
    set((s) => {
      const tab = s.tabs[tabId];
      if (!tab) return s;
      const seq = ++frameSeq;
      const frames = [...tab.frames, { ...frame, seq }];
      if (frames.length > MESSAGE_BUFFER_CAP) {
        frames.splice(0, frames.length - MESSAGE_BUFFER_CAP);
      }
      return { tabs: { ...s.tabs, [tabId]: { ...tab, frames } } };
    });
  },
}));
