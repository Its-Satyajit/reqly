export interface MonitorPoint {
	at: string;
	ok: boolean;
	status: number;
	latencyMs: number;
}

export interface MonitorAdapter {
	check(specPath: string): Promise<MonitorPoint>;
}

export const fallbackMonitorAdapter: MonitorAdapter = {
	async check() {
		throw new Error("monitor is not available in this build");
	},
};

let monitorBridge: MonitorAdapter | null = null;
export function setMonitorBridge(a: MonitorAdapter): void {
	monitorBridge = a;
}
export function getMonitorBridge(): MonitorAdapter {
	return monitorBridge ?? fallbackMonitorAdapter;
}
