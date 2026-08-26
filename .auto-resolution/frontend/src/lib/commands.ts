/** A command the palette can search and run. Extensible: any feature may add entries. */
export interface Command {
  id: string
  label: string
  group: 'Navigation' | 'Theme' | 'Environment' | 'Git' | 'Settings'
  keywords?: string[]
  run: () => void
}

/**
 * Commands derived from app state at render time. Search/filtering is owned by
 * the cmdk primitive in the palette UI. Git actions register into this same
 * list once their bindings exist (M44 T4/T7).
 */
export function buildCommands(sources: {
  views: readonly { id: string; label: string }[]
  navigate: (viewId: string) => void
  themes: readonly { id: string; label: string }[]
  setTheme: (themeId: string) => void
  environments: readonly { id: string; name: string }[]
  selectEnvironment: (envId: string) => void
}): Command[] {
  const cmds: Command[] = []

  for (const view of sources.views) {
    cmds.push({
      id: `nav.${view.id}`,
      label: `Go to ${view.label}`,
      group: 'Navigation',
      keywords: ['view', 'open'],
      run: () => sources.navigate(view.id),
    })
  }
  for (const theme of sources.themes) {
    cmds.push({
      id: `theme.${theme.id}`,
      label: `Theme: ${theme.label}`,
      group: 'Theme',
      keywords: ['appearance'],
      run: () => sources.setTheme(theme.id),
    })
  }
  for (const env of sources.environments) {
    cmds.push({
      id: `env.${env.id}`,
      label: `Environment: ${env.name}`,
      group: 'Environment',
      keywords: ['switch'],
      run: () => sources.selectEnvironment(env.id),
    })
  }
  return cmds
}

const GROUP_ORDER: Command['group'][] = [
  'Navigation',
  'Git',
  'Environment',
  'Theme',
  'Settings',
]

export function groupCommands(commands: Command[]): Map<Command['group'], Command[]> {
  const grouped = new Map<Command['group'], Command[]>()
  for (const group of GROUP_ORDER) {
    const members = commands.filter((c) => c.group === group)
    if (members.length > 0) grouped.set(group, members)
  }
  return grouped
}
