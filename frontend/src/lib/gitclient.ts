// Git status client shared by every Reqly front-end. The desktop host injects
// a GitAdapter backed by the Go core; browser dev mode uses the fallback.

export interface GitFileStatus {
	path: string;
	/** Index (staged) code, ' ' when unchanged. */
	x: string;
	/** Worktree code, ' ' when unchanged. */
	y: string;
	staged: boolean;
}

export interface GitStatusResult {
	branch: string;
	files: GitFileStatus[];
	clean: boolean;
	repoFound: boolean;
}

export interface GitAdapter {
	status(): Promise<GitStatusResult | null>;
	stage(paths: string[]): Promise<void>;
	unstage(paths: string[]): Promise<void>;
	commit(message: string): Promise<void>;
}

export const fallbackGitAdapter: GitAdapter = {
	async status() {
		return { branch: '', files: [], clean: true, repoFound: false };
	},
	async stage() {
		throw new Error('Git staging is only available in the Reqly desktop app');
	},
	async unstage() {
		throw new Error('Git staging is only available in the Reqly desktop app');
	},
	async commit() {
		throw new Error('Git commits are only available in the Reqly desktop app');
	},
};
