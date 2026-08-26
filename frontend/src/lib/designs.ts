export type DesignId =
  | 'current'
  | 'ember'
  | 'forge'
  | 'blueprint'
  | 'signal'
  | 'paper'

export interface DesignDef {
  id: DesignId
  label: string
}

/**
 * Open registry: a new design is a `[data-design='<id>']` block in
 * frontend/src/styles/designs/<id>.css plus one entry here. The axis is
 * purely presentational — application state lives in stores and survives
 * switching.
 *
 * Six candidates ship: the incumbent (current) plus five bold brand
 * designs. One will be chosen and improved; the rest removed.
 */
export const DESIGNS: readonly DesignDef[] = [
  { id: 'current', label: 'Current' },
  { id: 'ember', label: 'Ember' },
  { id: 'forge', label: 'Forge' },
  { id: 'blueprint', label: 'Blueprint' },
  { id: 'signal', label: 'Signal' },
  { id: 'paper', label: 'Paper' },
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
