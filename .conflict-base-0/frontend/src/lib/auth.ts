// Auth types + adapter shared by every Reqly front-end.
//
// Like request execution, auth never imports the Wails-generated bindings
// directly (they live in the host app and are regenerated on build). The host
// injects an AuthAdapter backed by the Go core's AuthService; browser dev
// mode uses a fallback that explains auth needs the desktop bridge.

export interface AuthTokenStatus {
  endpoint: string
  grantType: string
  expiry: string
  accessToken: string
  hasRefresh: boolean
  state: string
}

export interface AuthStatusData {
  backend: string
  tokens: AuthTokenStatus[]
}

export interface AuthLoginResult {
  accessToken: string
}

export interface AuthAdapter {
  login: (config: Record<string, string>, flow: string) => Promise<AuthLoginResult>
  status: () => Promise<AuthStatusData>
  logout: () => Promise<number>
}

/**
 * fallbackAuthAdapter is used in plain Vite dev mode (no Wails bridge): it
 * reports an empty store and rejects actions with an actionable message.
 */
export const fallbackAuthAdapter: AuthAdapter = {
  login: async () => {
    throw new Error('Auth login is only available in the desktop app')
  },
  status: async () => ({ backend: '-', tokens: [] }),
  logout: async () => 0,
}
