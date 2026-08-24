export type RealtimeKind = "ws" | "sse";
export type RealtimeStatus = "idle" | "connecting" | "connected" | "closed" | "error";

export interface RealtimeFrameView {
  /** Monotonic per-store id; stable React key even after buffer trimming. */
  seq?: number;
  type: "message" | "status" | "error" | "closed";
  direction?: "in" | "out";
  data?: string;
  encoding?: string;
  name?: string;
  id?: string;
  /** SSE retry hint in milliseconds ("retry:" field), when the server sent one. */
  retryMs?: number;
  timestamp: number;
}

export interface RealtimeHeaderPair {
  key: string;
  value: string;
}

export interface RealtimeAdapter {
  open(input: {
    sessionId: string;
    kind: RealtimeKind;
    url: string;
    headers?: RealtimeHeaderPair[];
  }): Promise<void>;
  send(sessionId: string, data: string): Promise<void>;
  /** sendBinary writes raw bytes (base64-encoded payload) as a WS binary frame. */
  sendBinary?(sessionId: string, base64: string): Promise<void>;
  close(sessionId: string): Promise<void>;
  subscribe(
    sessionId: string,
    onFrame: (frame: RealtimeFrameView) => void,
  ): () => void;
}

export const fallbackRealtimeAdapter: RealtimeAdapter = {
  async open() {
    throw new Error("realtime client is not available in this build");
  },
  async send() {
    throw new Error("realtime client is not available in this build");
  },
  async close() {
    /* nothing to close */
  },
  subscribe() {
    return () => {};
  },
};

/** MESSAGE_BUFFER_CAP bounds the inspector history (G-7.2.4 / G-7.1.4). */
export const MESSAGE_BUFFER_CAP = 500;

export function formatFrameTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString(undefined, { hour12: false }) +
    "." + String(d.getMilliseconds()).padStart(3, "0");
}
