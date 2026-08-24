import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
	fallbackGitAdapter,
	type GitAdapter,
	type GitStatusResult,
} from '#lib/gitclient'
import { useGitStore } from './useGitStore'

const noopAdapter = {
	status: async () => null,
	stage: async () => {},
	unstage: async () => {},
	commit: async () => {},
	worktrees: async () => [],
	addWorktree: async () => {},
	removeWorktree: async () => {},
	recentCommits: async () => [],
	conflicts: async () => [],
	resolveSide: async () => {},
	mergeAbort: async () => {},
}

function adapterWith(status: Partial<GitStatusResult> | null): GitAdapter {
	return {
		...noopAdapter,
		stage: vi.fn(async () => {}),
		unstage: vi.fn(async () => {}),
		commit: vi.fn(async () => {}),
		status: vi.fn(async () =>
			status === null
				? null
				: { branch: '', files: [], clean: true, repoFound: true, ...status },
		),
	}
}

describe('git store', () => {
	beforeEach(() => {
		useGitStore.getState().setAdapter(fallbackGitAdapter)
	})

	it('starts idle with the fallback adapter', () => {
		const s = useGitStore.getState()
		expect(s.status).toBeNull()
		expect(s.loading).toBe(false)
		expect(s.error).toBeNull()
	})

	it('refresh pulls status through the adapter', async () => {
		const adapter = adapterWith({ branch: 'main', clean: true })
		useGitStore.getState().setAdapter(adapter)
		await useGitStore.getState().refresh()
		expect(adapter.status).toHaveBeenCalledOnce()
		expect(useGitStore.getState().status?.branch).toBe('main')
		expect(useGitStore.getState().loading).toBe(false)
	})

	it('refresh surfaces errors without losing previous state', async () => {
		await useGitStore
			.getState()
			.setAdapterAndRefresh(adapterWith({ branch: 'main' }))
		const failing: GitAdapter = {
			...fallbackGitAdapter,
			status: vi.fn(async () => {
				throw new Error('boom')
			}),
		}
		useGitStore.getState().setAdapter(failing)
		await useGitStore.getState().refresh()
		expect(useGitStore.getState().error).toContain('boom')
		expect(useGitStore.getState().status?.branch).toBe('main')
	})

	it('stage and commit refresh afterwards', async () => {
		const adapter = adapterWith({
			files: [{ path: 'a.json', x: 'M', y: ' ', staged: true }],
			clean: false,
		})
		useGitStore.getState().setAdapter(adapter)
		await useGitStore.getState().stage(['a.json'])
		expect(adapter.stage).toHaveBeenCalledWith(['a.json'])
		expect(useGitStore.getState().status?.files[0]?.path).toBe('a.json')

		await useGitStore.getState().commit('feat: x')
		expect(adapter.commit).toHaveBeenCalledWith('feat: x')
	})

	it('setAdapterAndRefresh wires the host adapter in one call', async () => {
		const adapter = adapterWith({ branch: 'feature' })
		await useGitStore.getState().setAdapterAndRefresh(adapter)
		expect(useGitStore.getState().adapter).toBe(adapter)
		expect(useGitStore.getState().status?.branch).toBe('feature')
	})
})
