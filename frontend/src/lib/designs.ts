export type DesignId =
  | 'current'
  | 'modern'
  | 'ide'
  | 'inspector'
  | 'minimal'
  | 'command-center'

export interface DesignDef {
  id: DesignId
  label: string
}

/**
 * Open registry: a new design is a `[data-design='<id>']` block in
 * frontend/src/styles/designs.css plus one entry here. The axis is purely
 * presentational — application state lives in stores and survives switching.
 */
export const DESIGNS: readonly DesignDef[] = [
  { id: 'current', label: 'Current' },
  { id: 'modern', label: 'Modern' },
  { id: 'ide', label: 'IDE' },
  { id: 'inspector', label: 'Inspector' },
  { id: 'minimal', label: 'Minimal' },
  { id: 'command-center', label: 'Command Center' },
] as const

export const DEFAULT_DESIGN: DesignId = 'current'

function isDesignId(value: string): value is DesignId {
  return DESIGNS.some((d) => d.id === value)
}

export function resolveDesign(preference: string | null): DesignId {
  return preference !== null && isDesignId(preference) ? preference : DEFAULT_DESIGN
}

export function designById(id: DesignId): DesignDef {
  return DESIGNS.find((d) => d.id === id) ?? DESIGNS[0]
}
