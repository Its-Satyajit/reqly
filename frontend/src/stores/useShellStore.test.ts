import { beforeEach, describe, expect, it } from 'vitest'

import { initialShellState, useShellStore } from './useShellStore'

describe('shell store (inspector mount point)', () => {
  beforeEach(() => {
    useShellStore.getState().closeInspector()
    localStorage.clear()
  })

  it('starts with the inspector closed', () => {
    expect(useShellStore.getState().inspectorOpen).toBe(false)
    expect(useShellStore.getState().inspectorTab).toBeNull()
  })

  it('openInspector opens with a tab and remembers the last tab', () => {
    useShellStore.getState().openInspector('response-headers')
    expect(useShellStore.getState().inspectorOpen).toBe(true)
    expect(useShellStore.getState().inspectorTab).toBe('response-headers')
    useShellStore.getState().closeInspector()
    useShellStore.getState().openInspector()
    expect(useShellStore.getState().inspectorTab).toBe('response-headers')
  })

  it('toggle flips open state without losing the tab', () => {
    const { toggleInspector } = useShellStore.getState()
    toggleInspector()
    expect(useShellStore.getState().inspectorOpen).toBe(true)
    toggleInspector()
    expect(useShellStore.getState().inspectorOpen).toBe(false)
    expect(useShellStore.getState().inspectorTab).not.toBeNull()
  })

  it('persists the open flag and last tab', () => {
    useShellStore.getState().openInspector('timeline')
    expect(localStorage.getItem('reqly-shell-inspector-open')).toBe('1')
    expect(localStorage.getItem('reqly-shell-inspector-tab')).toBe('timeline')
  })

  it('restores persisted state on init', () => {
    localStorage.setItem('reqly-shell-inspector-open', '1')
    localStorage.setItem('reqly-shell-inspector-tab', 'commit-strip')
    // The store seeds its initial state through this factory.
    const fresh = initialShellState()
    expect(fresh.inspectorOpen).toBe(true)
    expect(fresh.inspectorTab).toBe('commit-strip')
  })
})
