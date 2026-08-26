export interface WorkspaceStatus {
  found: boolean;
  path?: string;
}

export interface WorkspaceBootstrapAdapter {
  status(): Promise<WorkspaceStatus>;
  restoreLast(): Promise<WorkspaceStatus>;
  pickFolder(): Promise<string>;
  open(dir: string): Promise<void>;
  create(dir: string, name?: string): Promise<void>;
}

export const fallbackWorkspaceBootstrapAdapter: WorkspaceBootstrapAdapter = {
  async status() {
    return { found: false };
  },
  async restoreLast() {
    return { found: false };
  },
  async pickFolder() {
    throw new Error("workspace picker is not available in this build");
  },
  async open() {
    throw new Error("workspace switching is not available in this build");
  },
  async create() {
    throw new Error("workspace creation is not available in this build");
  },
};

export const CREATE_HINT = "is not a Reqly workspace";

const RECENTS_KEY = "reqly.recentWorkspaces";
const RECENTS_MAX = 5;

/** readRecents loads the remembered workspace folders, newest first. */
export function readRecents(): string[] {
	try {
		const raw = localStorage.getItem(RECENTS_KEY);
		const parsed: unknown = raw ? JSON.parse(raw) : [];
		return Array.isArray(parsed)
			? parsed
					.filter((p): p is string => typeof p === "string")
					.slice(0, RECENTS_MAX)
			: [];
	} catch {
		return [];
	}
}

/** rememberWorkspace appends dir to the recents list, deduped, newest first. */
export function rememberWorkspace(dir: string): void {
	const next = [dir, ...readRecents().filter((p) => p !== dir)].slice(
		0,
		RECENTS_MAX,
	);
	try {
		localStorage.setItem(RECENTS_KEY, JSON.stringify(next));
	} catch {
		// Storage unavailable (e.g. restricted webview) — recents are a
		// convenience, not a requirement.
	}
}
