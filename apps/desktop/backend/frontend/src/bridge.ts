import { AppService } from '../bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/index'
import type {
  AuthAdapter,
  RequestInput,
  RequestSender,
  ResponseData,
} from '@reqly/frontend'
import { serializeBody } from '@reqly/frontend'
import { useAuthStore, useRequestStore } from '@reqly/frontend'

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
      tokens: status.tokens.map((t) => ({
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
 * Wires the Go core behind the shared request and auth stores. Called once
 * from the host entry point, before the React tree mounts.
 */
export function initRequestBridge(): void {
  useRequestStore.getState().setSender(wailsSender)
  useAuthStore.getState().setAdapter(wailsAuthAdapter)
}
