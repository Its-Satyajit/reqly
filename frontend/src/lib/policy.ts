export interface Policy {
	requireAudit: boolean;
	maxWorkflowSteps: number;
	allowedActions?: string[] | null;
	requireAuth?: boolean;
	allowCustomThemes?: boolean;
}

export interface PolicyAdapter {
	get(): Promise<Policy>;
	save(policy: Policy): Promise<void>;
	enforce(action: string, resource: string): Promise<void>;
}

export const fallbackPolicyAdapter: PolicyAdapter = {
	async get() {
		throw new Error("policy is not available in this build");
	},
	async save() {
		throw new Error("policy is not available in this build");
	},
	async enforce() {
		throw new Error("policy is not available in this build");
	},
};

let policyBridge: PolicyAdapter | null = null;
export function setPolicyBridge(a: PolicyAdapter): void {
	policyBridge = a;
}
export function getPolicyBridge(): PolicyAdapter {
	return policyBridge ?? fallbackPolicyAdapter;
}

export function validatePolicy(p: Policy): string | null {
	if (p.maxWorkflowSteps < 0) return "maxWorkflowSteps must be >= 0";
	if (p.allowedActions?.some((a) => a.trim() === "")) return "allowedActions cannot contain empty";
	return null;
}
