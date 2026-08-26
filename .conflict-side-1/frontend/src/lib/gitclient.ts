// Git status client shared by every Reqly front-end. The desktop host injects
// a GitAdapter backed by the Go core; browser dev mode uses the fallback.

export interface GitFileStatus {
	path: string;
	/** Index (staged) code, ' ' when unchanged. */
	x: string;
	/** Worktree code, ' ' when unchanged. */
	y: string;
	staged: boolean;
	/** Changed-line counts vs HEAD (0 for untracked files). */
	adds: number;
	dels: number;
}

export interface GitStatusResult {
	branch: string;
	files: GitFileStatus[];
	clean: boolean;
	repoFound: boolean;
}

export interface GitWorktree {
	path: string;
	branch: string;
	isCurrent: boolean;
	isBare: boolean;
	detached: boolean;
}

export interface GitRecentCommit {
	hash: string;
	subject: string;
}

export interface GitConflictFile {
	path: string;
	code: string;
}

export interface GitAdapter {
	status(): Promise<GitStatusResult | null>;
	stage(paths: string[]): Promise<void>;
	unstage(paths: string[]): Promise<void>;
	commit(message: string): Promise<void>;
	worktrees(): Promise<GitWorktree[]>;
	addWorktree(path: string): Promise<void>;
	removeWorktree(path: string): Promise<void>;
	recentCommits(limit: number): Promise<GitRecentCommit[]>;
	conflicts(): Promise<GitConflictFile[]>;
	resolveSide(path: string, side: 'ours' | 'theirs'): Promise<void>;
	mergeAbort(): Promise<void>;
}

const desktopOnly = (what: string) => async () => {
	throw new Error(`${what} is only available in the Reqly desktop app`);
};

export const fallbackGitAdapter: GitAdapter = {
	async status() {
		return { branch: '', files: [], clean: true, repoFound: false };
	},
	stage: desktopOnly('Staging'),
	unstage: desktopOnly('Unstaging'),
	commit: desktopOnly('Committing'),
	worktrees: async () => [],
	addWorktree: desktopOnly('Worktree creation'),
	removeWorktree: desktopOnly('Worktree removal'),
	recentCommits: async () => [],
	conflicts: async () => [],
	resolveSide: desktopOnly('Conflict resolution'),
	mergeAbort: desktopOnly('Merge abort'),
};
