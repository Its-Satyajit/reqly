import { create } from 'zustand'

import {
	fallbackGitAdapter,
	type GitAdapter,
	type GitStatusResult,
} from '#lib/gitclient'

interface GitState {
	adapter: GitAdapter
	status: GitStatusResult | null
	loading: boolean
	error: string | null
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
			const status = await get().adapter.status()
			set({ status, error: null })
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
