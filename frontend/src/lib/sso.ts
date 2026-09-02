export interface SSOConfig {
	issuer: string;
	clientId: string;
	jwksUrl?: string;
	allowedGroups?: string[];
}

export interface SCIMUser {
	id: string;
	userName: string;
	email?: string;
	groups?: string[];
	active: boolean;
}

export interface SSOAdapter {
	validate(issuer: string, clientId: string, token: string, secret: string): Promise<void>;
}

export interface SCIMAdapter {
	createUser(userName: string, email: string): Promise<SCIMUser>;
	listUsers(): Promise<SCIMUser[]>;
}

export const fallbackSSOAdapter: SSOAdapter = {
	async validate() {
		throw new Error("sso is not available in this build");
	},
};

export const fallbackSCIMAdapter: SCIMAdapter = {
	async createUser() {
		throw new Error("scim is not available in this build");
	},
	async listUsers() {
		throw new Error("scim is not available in this build");
	},
};

let ssoBridge: SSOAdapter | null = null;
let scimBridge: SCIMAdapter | null = null;

export function setSSOBridge(a: SSOAdapter): void {
	ssoBridge = a;
}
export function getSSOBridge(): SSOAdapter {
	return ssoBridge ?? fallbackSSOAdapter;
}
export function setSCIMBridge(a: SCIMAdapter): void {
	scimBridge = a;
}
export function getSCIMBridge(): SCIMAdapter {
	return scimBridge ?? fallbackSCIMAdapter;
}
