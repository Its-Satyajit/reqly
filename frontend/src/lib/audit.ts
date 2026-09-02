export interface AuditEntry {
	id: string;
	timestamp: string;
	actor: string;
	action: string;
	resource: string;
	details?: string;
}

export interface AuditAdapter {
	list(): Promise<AuditEntry[]>;
	add(action: string, resource: string, details: string): Promise<AuditEntry>;
	clear(): Promise<void>;
	export(): Promise<string>;
}

export const fallbackAuditAdapter: AuditAdapter = {
	async list() {
		throw new Error("audit is not available in this build");
	},
	async add() {
		throw new Error("audit is not available in this build");
	},
	async clear() {
		throw new Error("audit is not available in this build");
	},
	async export() {
		throw new Error("audit is not available in this build");
	},
};

let auditBridge: AuditAdapter | null = null;
export function setAuditBridge(a: AuditAdapter): void {
	auditBridge = a;
}
export function getAuditBridge(): AuditAdapter {
	return auditBridge ?? fallbackAuditAdapter;
}

export function formatAuditTime(iso: string): string {
	try {
		const d = new Date(iso);
		return d.toLocaleString(undefined, {
			year: "numeric",
			month: "short",
			day: "2-digit",
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
			hour12: false,
		});
	} catch {
		return iso;
	}
}
