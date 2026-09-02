export interface MqttPublishInput {
	broker: string;
	topic: string;
	message: string;
	qos: number;
	retain: boolean;
	username?: string;
	password?: string;
	clientId?: string;
}

export interface MqttSubscribeInput {
	sessionId: string;
	broker: string;
	topic: string;
	qos: number;
	username?: string;
	password?: string;
	clientId?: string;
}

export interface MqttFrameView {
	sessionId: string;
	type: "message" | "status" | "error" | "closed";
	topic?: string;
	payload?: string;
	qos?: number;
	retain?: boolean;
	data?: string;
	timestamp: number;
}

export interface MqttAdapter {
	publish(input: MqttPublishInput): Promise<void>;
	subscribe(
		input: MqttSubscribeInput,
	): Promise<void>;
	cancel(sessionId: string): Promise<void>;
	onFrame(
		sessionId: string,
		cb: (frame: MqttFrameView) => void,
	): () => void;
}

export const fallbackMqttAdapter: MqttAdapter = {
	async publish() {
		throw new Error("mqtt client is not available in this build");
	},
	async subscribe() {
		throw new Error("mqtt client is not available in this build");
	},
	async cancel() {},
	onFrame() {
		return () => {};
	},
};

let mqttBridge: MqttAdapter | null = null;

export function setMqttBridge(adapter: MqttAdapter): void {
	mqttBridge = adapter;
}

export function getMqttBridge(): MqttAdapter {
	return mqttBridge ?? fallbackMqttAdapter;
}

export function qosLabel(qos: number): string {
	switch (qos) {
		case 0: return "QoS 0";
		case 1: return "QoS 1";
		case 2: return "QoS 2";
		default: return `QoS ${qos}`;
	}
}

export function qosTint(qos: number): string {
	switch (qos) {
		case 1: return "text-status-warn border-status-warn/30 bg-status-warn/10";
		case 2: return "text-status-error border-status-error/30 bg-status-error/10";
		default: return "text-muted-foreground border-border bg-muted/30";
	}
}
