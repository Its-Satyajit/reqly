export interface Role {
	name: string;
	permissions: string[];
}

export interface RBAC {
	roles: Record<string, Role>;
	userRoles: Record<string, string>;
}

export interface RBACAdapter {
	get(): Promise<RBAC>;
	listRoles(): Promise<string[]>;
	check(user: string, action: string, resource: string): Promise<void>;
}

export const fallbackRBACAdapter: RBACAdapter = {
	async get() {
		throw new Error("rbac is not available in this build");
	},
	async listRoles() {
		throw new Error("rbac is not available in this build");
	},
	async check() {
		throw new Error("rbac is not available in this build");
	},
};

let rbacBridge: RBACAdapter | null = null;
export function setRBACBridge(a: RBACAdapter): void {
	rbacBridge = a;
}
export function getRBACBridge(): RBACAdapter {
	return rbacBridge ?? fallbackRBACAdapter;
}

export const ALL_PERMISSIONS = [
	"request.send",
	"workflow.run",
	"automation.run",
	"collection.run",
	"theme.import",
	"theme.export",
	"auth.login",
	"mock.start",
] as const;

export function canRole(role: Role, action: string): boolean {
	return role.permissions.includes("*") || role.permissions.includes(action);
}
