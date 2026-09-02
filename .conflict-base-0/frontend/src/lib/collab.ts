export interface Collaborator {
	user: string;
	role: string;
	addedAt: string;
}

export interface SharedWorkspace {
	path: string;
	collaborators: Collaborator[];
}

export interface CollabAdapter {
	list(): Promise<Collaborator[]>;
	add(user: string, role: string): Promise<void>;
	remove(user: string): Promise<void>;
	serve(port: number): Promise<string>;
}

export const fallbackCollabAdapter: CollabAdapter = {
	async list() {
		throw new Error("collab is not available in this build");
	},
	async add() {
		throw new Error("collab is not available in this build");
	},
	async remove() {
		throw new Error("collab is not available in this build");
	},
	async serve() {
		throw new Error("collab is not available in this build");
	},
};

let collabBridge: CollabAdapter | null = null;
export function setCollabBridge(a: CollabAdapter): void {
	collabBridge = a;
}
export function getCollabBridge(): CollabAdapter {
	return collabBridge ?? fallbackCollabAdapter;
}

export const COLLAB_ROLES = [
	{ value: "viewer", label: "Viewer — can send & export" },
	{ value: "editor", label: "Editor — can run workflows & import" },
	{ value: "admin", label: "Admin — all permissions" },
] as const;

export function roleBadgeClass(role: string): string {
	switch (role) {
		case "admin":
			return "bg-status-warn/10 text-status-warn border-status-warn/20";
		case "editor":
			return "bg-status-redirect/10 text-status-redirect border-status-redirect/20";
		default:
			return "bg-muted text-muted-foreground border-border";
	}
}
