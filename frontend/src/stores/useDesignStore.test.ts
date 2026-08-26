import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { DEFAULT_DESIGN } from '#lib/designs'

describe('useDesignStore', () => {
  let localStorageGet: ReturnType<typeof vi.spyOn> | undefined
  let localStorageSet: ReturnType<typeof vi.spyOn> | undefined

  const load = async () => {
    const mod = await import('./useDesignStore')
    return mod.useDesignStore
  }

  beforeEach(() => {
    vi.resetModules()
    localStorageGet = vi.spyOn(Storage.prototype, 'getItem')
    localStorageSet = vi.spyOn(Storage.prototype, 'setItem')
    document.documentElement.dataset.design = ''
  })

  afterEach(() => {
    localStorageGet?.mockRestore()
    localStorageSet?.mockRestore()
  })

  it('defaults to the current design when nothing is stored', async () => {
    localStorageGet!.mockReturnValue(null)
    const store = await load()
    expect(store.getState().design).toBe(DEFAULT_DESIGN)
    expect(document.documentElement.dataset.design).toBe('current')
  })

  it('restores a stored design and applies it to the DOM', async () => {
    localStorageGet!.mockReturnValue('ember')
    const store = await load()
    expect(store.getState().design).toBe('ember')
    expect(document.documentElement.dataset.design).toBe('ember')
  })

  it('falls back to the default for unknown stored values', async () => {
    localStorageGet!.mockReturnValue('neon')
    const store = await load()
    expect(store.getState().design).toBe(DEFAULT_DESIGN)
  })

  it('setDesign updates state, DOM attribute, and persistence', async () => {
    localStorageGet!.mockReturnValue(null)
    const store = await load()
    store.getState().setDesign('ember')
    expect(store.getState().design).toBe('ember')
    expect(store.getState().label).toBe('Ember')
    expect(document.documentElement.dataset.design).toBe('ember')
    expect(localStorageSet).toHaveBeenCalledWith('reqly-design', 'ember')
  })
})
