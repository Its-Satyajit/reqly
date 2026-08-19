import { AppService } from '../bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/index'
import type {
  AuthAdapter,
  CollectionsAdapter,
  EnvAdapter,
  RequestInput,
  RequestSender,
  ResponseData,
} from '@reqly/frontend'
import { serializeBody } from '@reqly/frontend'
import { useAuthStore, useRequestStore, useWorkspaceStore } from '@reqly/frontend'

/**
 * wailsSender executes requests through the Go core via the generated Wails
 * bindings, then normalizes the core response into the shared ResponseData
 * shape the UI renders.
 */
export const wailsSender: RequestSender = async (req: RequestInput): Promise<ResponseData> => {
  const headers = (req.headers ?? []).map(({ key, value }) => ({ key, value }))
  const { body, contentType } = serializeBody(req)
  const hasManualType = headers.some(
    (h) => h.key.toLowerCase() === 'content-type',
  )
  if (contentType && !hasManualType) headers.push({ key: 'Content-Type', value: contentType })

  const res = await AppService.SendRequest({
    method: req.method,
    url: req.url,
    headers,
    query: (req.params ?? []).map(({ key, value }) => ({ key, value })),
    body,
  } as never)
  if (!res) {
    throw new Error('core returned an empty response')
  }
  return res as ResponseData
}

/**
 * wailsAuthAdapter executes auth actions through the Go core's AuthService
 * via the generated Wails bindings, shaping the results into the shared
 * AuthAdapter contract the auth panel renders.
 */
export const wailsAuthAdapter: AuthAdapter = {
  login: async (config, flow) => {
    const tok = await AppService.AuthLogin(config, flow)
    if (!tok) {
      throw new Error('login returned no token')
    }
    return { accessToken: tok.AccessToken }
  },
  status: async () => {
    const status = await AppService.AuthStatus()
    if (!status) {
      throw new Error('core returned an empty status')
    }
    return {
      backend: status.backend,
      tokens: (status.tokens ?? []).map((t) => ({
        endpoint: t.endpoint,
        grantType: t.grantType,
        expiry: t.expiry,
        accessToken: t.accessToken,
        hasRefresh: t.hasRefresh,
        state: t.state,
      })),
    }
  },
  logout: async () => AppService.AuthLogout(),
}

/**
 * normalizeVariables coerces the generated bindings' nullable/undefined-valued
 * map into a plain string map the shared types expect.
 */
const normalizeVariables = (v: Record<string, string | undefined> | null | undefined): Record<string, string> => {
  const out: Record<string, string> = {}
  for (const [k, val] of Object.entries(v ?? {})) {
    if (typeof val === 'string') out[k] = val
  }
  return out
}

export const wailsEnvAdapter: EnvAdapter = {
  list: async () => {
    const data = await AppService.EnvList()
    if (!data) {
      throw new Error('core returned an empty environment list')
    }
    return {
      active: data.active ?? '',
      environments: (data.environments ?? []).map((e) => ({
        name: e.name,
        description: e.description ?? '',
        variables: normalizeVariables(e.variables),
        secrets: e.secrets ?? [],
      })),
    }
  },
  read: async (name) => {
    const env = await AppService.EnvRead(name)
    if (!env) {
      throw new Error(`core returned an empty environment "${name}"`)
    }
    return {
      name: env.name,
      description: env.description ?? '',
      variables: normalizeVariables(env.variables),
      secrets: env.secrets ?? [],
    }
  },
  setActive: async (name) => {
    await AppService.EnvSetActive(name)
  },
  create: async (name, description, variables) => {
    await AppService.EnvCreate(name, description, variables)
  },
  update: async (name, description, variables) => {
    await AppService.EnvUpdate(name, description, variables)
  },
  updateSecrets: async (name, values, remove) => {
    await AppService.EnvUpdateSecrets(name, values, remove)
  },
  delete: async (name) => {
    await AppService.EnvDelete(name)
  },
}

/**
 * wailsCollectionsAdapter loads the workspace's collection tree through the
 * Go core's WorkspaceService via the generated Wails bindings. The generated
 * models are nullable; normalize them to the shared tree shapes.
 */
type WailsTree = NonNullable<Awaited<ReturnType<typeof AppService.WorkspaceLoad>>>
type WailsCollection = NonNullable<WailsTree['collections']>[number]
type WailsFolder = NonNullable<WailsCollection['folders']>[number]
type WailsRequest = NonNullable<WailsCollection['requests']>[number]

const normalizeFolder = (f: WailsFolder): import('@reqly/frontend').WorkspaceFolder => ({
  name: f.name,
  path: f.path,
  folders: (f.folders ?? []).map(normalizeFolder),
  requests: (f.requests ?? []).map(normalizeRequest),
})

const normalizeRequest = (r: WailsRequest): import('@reqly/frontend').WorkspaceRequest => ({
  name: r.name,
  path: r.path,
})

export const wailsCollectionsAdapter: CollectionsAdapter = {
  load: async () => {
    const tree = await AppService.WorkspaceLoad()
    if (!tree) {
      throw new Error('core returned an empty workspace tree')
    }
    return {
      name: tree.name ?? '',
      path: tree.path ?? '',
      collections: (tree.collections ?? []).map((c) => ({
        name: c.name,
        path: c.path,
        folders: (c.folders ?? []).map(normalizeFolder),
        requests: (c.requests ?? []).map(normalizeRequest),
      })),
    }
  },
}

/**
 * Wires the Go core behind the shared request, auth, and environment stores.
 * Called once from the host entry point, before the React tree mounts.
 */
export function initRequestBridge(): void {
  useRequestStore.getState().setSender(wailsSender)
  useAuthStore.getState().setAdapter(wailsAuthAdapter)
  useWorkspaceStore.getState().setEnvAdapter(wailsEnvAdapter)
  useWorkspaceStore.getState().setWorkspaceAdapter(wailsCollectionsAdapter)
}
