export interface SocketIOConnectInput {
	sessionId: string;
	url: string;
	namespace?: string;
}

export interface SocketIOEmitInput {
	sessionId: string;
	url: string;
	event: string;
	data?: unknown;
	namespace?: string;
}

export interface SocketIOFrameView {
	sessionId: string;
	type: "message" | "status" | "error" | "closed";
	namespace?: string;
	event?: string;
	data?: unknown;
	raw?: string;
	timestamp: number;
}

export interface SocketIOAdapter {
	connect(input: SocketIOConnectInput): Promise<void>;
	emit(input: SocketIOEmitInput): Promise<void>;
	close(sessionId: string): Promise<void>;
	onFrame(
		sessionId: string,
		cb: (frame: SocketIOFrameView) => void,
	): () => void;
}

export const fallbackSocketIOAdapter: SocketIOAdapter = {
	async connect() {
		throw new Error("socket.io client is not available in this build");
	},
	async emit() {
		throw new Error("socket.io client is not available in this build");
	},
	async close() {},
	onFrame() {
		return () => {};
	},
};

let socketBridge: SocketIOAdapter | null = null;

export function setSocketIOBridge(adapter: SocketIOAdapter): void {
	socketBridge = adapter;
}

export function getSocketIOBridge(): SocketIOAdapter {
	return socketBridge ?? fallbackSocketIOAdapter;
}

export function formatSocketIOTime(ts: number): string {
	const d = new Date(ts);
	return d.toLocaleTimeString(undefined, { hour12: false }) + "." + String(d.getMilliseconds()).padStart(3, "0");
}
