import { beforeEach, describe, expect, it } from 'vitest'

import { usePaletteStore } from './usePaletteStore'

describe('palette store', () => {
  beforeEach(() => {
    usePaletteStore.getState().close()
    localStorage.clear()
  })

  it('starts closed', () => {
    expect(usePaletteStore.getState().open).toBe(false)
  })

  it('open and close toggle state', () => {
    usePaletteStore.getState().openPalette()
    expect(usePaletteStore.getState().open).toBe(true)
    usePaletteStore.getState().close()
    expect(usePaletteStore.getState().open).toBe(false)
  })

  it('toggle flips state', () => {
    const { toggle } = usePaletteStore.getState()
    toggle()
    expect(usePaletteStore.getState().open).toBe(true)
    toggle()
    expect(usePaletteStore.getState().open).toBe(false)
  })

  it('records recently run commands, most recent first, deduped', () => {
    const s = useShellRecorder()
    s.run('a')
    s.run('b')
    s.run('a')
    expect(usePaletteStore.getState().recent).toEqual(['a', 'b'])
  })

  it('caps the recent list', () => {
    const s = useShellRecorder()
    for (const id of ['1', '2', '3', '4', '5', '6']) s.run(id)
    expect(usePaletteStore.getState().recent.length).toBeLessThanOrEqual(5)
  })
})

function useShellRecorder() {
  return {
    run: (id: string) => usePaletteStore.getState().recordRun(id),
  }
}
