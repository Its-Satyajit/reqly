// Environment types + adapter shared by every Reqly front-end.
//
// Like request execution and auth, environments never import the Wails
// generated bindings directly (they live in the host app and are regenerated
// on build). The host injects an EnvAdapter backed by the Go core's
// EnvironmentService; browser dev mode uses a read-only fallback.

/** EnvironmentData is the bridge-friendly view of one environment. Secret
 * values are never present — only their names, so the UI can show masked rows. */
export interface EnvironmentData {
  name: string
  description: string
  variables: Record<string, string>
  secrets: string[]
}

/** EnvListData is the result of listing a workspace's environments plus the
 * currently active one. */
export interface EnvListData {
  active: string
  environments: EnvironmentData[]
}

/** EnvAdapter is the seam through which the desktop UI reads and writes
 * environments. Secret values are never read back over this interface. */
export interface EnvAdapter {
  list: () => Promise<EnvListData>
  read: (name: string) => Promise<EnvironmentData>
  create: (name: string, description: string, variables: Record<string, string>) => Promise<void>
  setActive: (name: string) => Promise<void>
}

/**
 * fallbackEnvAdapter is used in plain Vite dev mode (no Wails bridge): it
 * reports an empty environment set and rejects mutations with an actionable
 * message.
 */
export const fallbackEnvAdapter: EnvAdapter = {
  list: async () => ({ active: '', environments: [] }),
  read: async (name) => {
    throw new Error(`Environment "${name}" is only available in the desktop app`)
  },
  create: async () => {
    throw new Error('Creating an environment is only available in the desktop app')
  },
  setActive: async () => {
    throw new Error('Setting an active environment is only available in the desktop app')
  },
}