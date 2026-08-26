import { describe, expect, it } from 'vitest'

import { buildCommands } from './commands'

const sources = {
  views: [
    { id: 'history', label: 'History' },
    { id: 'grpc', label: 'gRPC client' },
  ],
  navigate: (id: string) => void id,
  themes: [{ id: 'atlas-light', label: 'Atlas Light' }],
  setTheme: (id: string) => void id,
  environments: [{ id: 'env1', name: 'staging' }],
  selectEnvironment: (id: string) => void id,
}

describe('buildCommands', () => {
  it('creates one navigation command per view', () => {
    const cmds = buildCommands(sources)
    const nav = cmds.filter((c) => c.group === 'Navigation')
    expect(nav.map((c) => c.id)).toEqual(['nav.history', 'nav.grpc'])
    expect(nav[0].label).toBe('Go to History')
  })

  it('creates theme and environment commands', () => {
    const cmds = buildCommands(sources)
    expect(cmds.find((c) => c.id === 'theme.atlas-light')?.label).toBe(
      'Theme: Atlas Light',
    )
    expect(cmds.find((c) => c.id === 'env.env1')?.label).toBe(
      'Environment: staging',
    )
  })

  it('wires run() through to the source callbacks', () => {
    const calls: string[] = []
    const cmds = buildCommands({
      ...sources,
      navigate: (id) => calls.push(`view:${id}`),
      setTheme: (id) => calls.push(`theme:${id}`),
      selectEnvironment: (id) => calls.push(`env:${id}`),
    })
    cmds.find((c) => c.id === 'nav.history')!.run()
    cmds.find((c) => c.id === 'theme.atlas-light')!.run()
    cmds.find((c) => c.id === 'env.env1')!.run()
    expect(calls).toEqual(['view:history', 'theme:atlas-light', 'env:env1'])
  })
})
