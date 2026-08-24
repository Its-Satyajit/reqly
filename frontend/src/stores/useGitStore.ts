import { create } from 'zustand'

import {
	fallbackGitAdapter,
	type GitAdapter,
	type GitConflictFile,
	type GitRecentCommit,
	type GitStatusResult,
	type GitWorktree,
} from '#lib/gitclient'

interface GitState {
	adapter: GitAdapter
	status: GitStatusResult | null
	worktrees: GitWorktree[]
	recentCommits: GitRecentCommit[]
	conflicts: GitConflictFile[]
	loading: boolean
	error: string | null
	resolveSide: (path: string, side: 'ours' | 'theirs') => Promise<boolean>
	mergeAbort: () => Promise<boolean>
	addWorktree: (path: string) => Promise<boolean>
	removeWorktree: (path: string) => Promise<boolean>
	setAdapter: (adapter: GitAdapter) => void
	setAdapterAndRefresh: (adapter: GitAdapter) => Promise<void>
	refresh: () => Promise<void>
	stage: (paths: string[]) => Promise<boolean>
	unstage: (paths: string[]) => Promise<boolean>
	commit: (message: string) => Promise<boolean>
}

/** Repository state for the git-native sidebar (M44 T4). */
export const useGitStore = create<GitState>()((set, get) => ({
	adapter: fallbackGitAdapter,
	status: null,
	worktrees: [],
	recentCommits: [],
	conflicts: [],
	loading: false,
	error: null,

	setAdapter(adapter) {
		set({ adapter, error: null })
	},

	async setAdapterAndRefresh(adapter) {
		get().setAdapter(adapter)
		await get().refresh()
	},

	async refresh() {
		set({ loading: true })
		try {
			const [status, worktrees, recentCommits, conflicts] = await Promise.all([
				get().adapter.status(),
				get().adapter.worktrees().catch(() => []),
				get().adapter.recentCommits(5).catch(() => []),
				get().adapter.conflicts().catch(() => []),
			])
			set({ status, worktrees, recentCommits, conflicts, error: null })
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
		} finally {
			set({ loading: false })
		}
	},

	async stage(paths) {
		try {
			await get().adapter.stage(paths)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async unstage(paths) {
		try {
			await get().adapter.unstage(paths)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async resolveSide(path, side) {
		try {
			await get().adapter.resolveSide(path, side)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async mergeAbort() {
		try {
			await get().adapter.mergeAbort()
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async addWorktree(path) {
		try {
			await get().adapter.addWorktree(path)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async removeWorktree(path) {
		try {
			await get().adapter.removeWorktree(path)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},

	async commit(messageText) {
		try {
			await get().adapter.commit(messageText)
			await get().refresh()
			return true
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) })
			return false
		}
	},
}))
