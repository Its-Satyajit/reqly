export interface GrpcMethod {
  name: string;
  fullName: string;
  inputType: string;
  outputType: string;
  serverStreaming: boolean;
}

export interface GrpcService {
  name: string;
  methods: GrpcMethod[];
}

export interface GrpcResultView {
  ok: boolean;
  messageJson?: string;
  codeName?: string;
  statusMessage?: string;
  durationMs?: number;
}

export interface GrpcStreamMessage {
  seq: number;
  messageJson: string;
}

export type GrpcStreamStatus = "idle" | "streaming" | "done" | "error" | "cancelled";

/** GrpcRequest is the request-file model the bridge evaluates through the
 * pipeline (env, variables, masking, history). */
export interface GrpcRequest {
  url: string;
  headers: { key: string; value: string }[];
  grpc: {
    service: string;
    method: string;
    message?: unknown;
    timeout?: string;
    protoFiles?: string[];
    tls?: boolean;
    tlsSkipVerify?: boolean;
    caFile?: string;
  };
}

export interface GrpcAdapter {
  services(input: { target: string; protoFiles?: string[] }): Promise<GrpcService[]>;
  invoke(sessionId: string, request: GrpcRequest): Promise<GrpcResultView>;
  stream(sessionId: string, request: GrpcRequest): Promise<void>;
  cancel(sessionId: string): Promise<void>;
  subscribe(
    sessionId: string,
    onEvent: (event: { type: "message" | "done" | "error" | "cancelled"; seq?: number; data?: string; codeName?: string }) => void,
  ): () => void;
}

export const fallbackGrpcAdapter: GrpcAdapter = {
  async services() {
    throw new Error("gRPC client is not available in this build");
  },
  async invoke() {
    throw new Error("gRPC client is not available in this build");
  },
  async stream() {
    throw new Error("gRPC client is not available in this build");
  },
  async cancel() {},
  subscribe() {
    return () => {};
  },
};
