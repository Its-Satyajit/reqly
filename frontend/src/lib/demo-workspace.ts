import type {
  CollectionsAdapter,
  OpenedRequest,
  WorkspaceTree,
} from './collections'
import type { EnvAdapter } from './env'
import type { HistoryAdapter, HistoryDetail } from './history'
import type { WorkspaceBootstrapAdapter } from './workspace'

/**
 * Dev-only demo adapters for `vite --?demo` design previews. They stand in for
 * the Wails bridge with a small in-memory workspace so the full shell —
 * sidebar, editor, environments, history — renders in a plain browser.
 * Never imported outside the `import.meta.env.DEV && ?demo` gate in main.tsx.
 */

const tree: WorkspaceTree = {
  name: 'reqly-demo',
  path: '/tmp/reqly-demo',
  collections: [
    {
      name: 'Payments API',
      path: 'payments',
      requests: [{ name: 'Service health', path: 'payments/health', method: 'GET' }],
      folders: [
        {
          name: 'Checkout',
          path: 'payments/checkout',
          requests: [
            { name: 'Create charge', path: 'payments/checkout/create-charge', method: 'POST' },
            { name: 'Get charge', path: 'payments/checkout/get-charge', method: 'GET' },
          ],
          folders: [
            {
              name: 'Refunds',
              path: 'payments/checkout/refunds',
              requests: [
                { name: 'Issue refund', path: 'payments/checkout/refunds/issue', method: 'POST' },
              ],
              folders: [],
            },
          ],
        },
      ],
    },
    {
      name: 'Users',
      path: 'users',
      requests: [
        { name: 'Login', path: 'users/login', method: 'POST' },
        { name: 'Current user', path: 'users/me', method: 'GET' },
      ],
      folders: [],
    },
  ],
}

const fileFor = (path: string): OpenedRequest['fileRequest'] => {
  const base = {
    headers: [{ key: 'Accept', value: 'application/json' }],
    query: [],
    body: '',
  }
  if (path.includes('create-charge') || path.includes('issue')) {
    return {
      method: 'POST',
      url: '{{baseUrl}}/v1/charges',
      ...base,
      body: '{\n  "amount": 1999,\n  "currency": "usd",\n  "source": "tok_visa"\n}',
    }
  }
  if (path === 'users/login') {
    return {
      method: 'POST',
      url: '{{baseUrl}}/login',
      headers: [{ key: 'Content-Type', value: 'application/json' }],
      query: [],
      body: '{\n  "email": "ada@example.com",\n  "password": "hunter2"\n}',
    }
  }
  return { method: 'GET', url: '{{baseUrl}}/v1/me', ...base }
}

const allRequests = (): { name: string; path: string; method?: string }[] => {
  const out: { name: string; path: string; method?: string }[] = []
  for (const c of tree.collections) {
    out.push(...c.requests)
    const walk = (f: WorkspaceTree['collections'][number]['folders'][number]) => {
      out.push(...f.requests)
      for (const child of f.folders) walk(child)
    }
    for (const f of c.folders) walk(f)
  }
  return out
}

const opened = (path: string): OpenedRequest => {
  const req = allRequests().find((r) => r.path === path)
  return {
    path,
    name: req?.name ?? path,
    request: {
      method: req?.method ?? 'GET',
      url: 'https://api.demo.reqly.dev/v1/me',
      headers: [{ key: 'Accept', value: 'application/json' }],
      query: [],
      body: '',
    },
    fileRequest: fileFor(path),
    variables: [
      { name: "baseUrl", value: "https://api.demo.reqly.dev", scope: "environment" },
    ],
    fileEnv: '',
    version: 'demo-1',
  }
}

export const demoCollectionsAdapter: CollectionsAdapter = {
  load: async () => tree,
  open: async (path) => opened(path),
  run: async () => {
    throw new Error('Collection runs are not available in the browser demo.')
  },
  cancelRun: async () => {},
  save: async (_path, draft) => JSON.stringify(draft),
  createCollection: async () => {},
  createFolder: async () => {},
  duplicateRequest: async (path) => `${path}-copy`,
}

const environments = [
  {
    name: 'dev',
    description: 'Local development',
    variables: { baseUrl: 'https://api.demo.reqly.dev' },
    secrets: ['api_key'],
  },
  {
    name: 'staging',
    description: 'Staging cluster',
    variables: { baseUrl: 'https://staging.demo.reqly.dev' },
    secrets: [],
  },
]

export const demoEnvAdapter: EnvAdapter = {
  list: async () => ({ active: 'dev', environments }),
  read: async (name) =>
    environments.find((e) => e.name === name) ?? environments[0],
  create: async () => {},
  update: async () => {},
  updateSecrets: async () => {},
  delete: async () => {},
  setActive: async () => {},
}

const historyEntries: HistoryDetail[] = [
  {
    id: 'h3',
    requestPath: 'payments/checkout/create-charge',
    method: 'POST',
    url: 'https://api.demo.reqly.dev/v1/charges',
    status: 201,
    durationMs: 214,
    size: 842,
    env: 'dev',
    createdAt: new Date(Date.now() - 1000 * 60 * 4).toISOString(),
    reqHeaders: { accept: ['application/json'] },
    reqBody: '{ "amount": 1999 }',
    respHeaders: { 'content-type': ['application/json'] },
    respBody: '{\n  "id": "ch_new",\n  "status": "succeeded"\n}',
  },
  {
    id: 'h2',
    requestPath: 'users/me',
    method: 'GET',
    url: 'https://api.demo.reqly.dev/v1/me',
    status: 401,
    durationMs: 96,
    size: 131,
    env: 'dev',
    createdAt: new Date(Date.now() - 1000 * 60 * 22).toISOString(),
    reqHeaders: { accept: ['application/json'] },
    reqBody: '',
    respHeaders: { 'content-type': ['application/json'] },
    respBody: '{\n  "error": "invalid_api_key"\n}',
  },
  {
    id: 'h1',
    requestPath: 'payments/checkout/get-charge',
    method: 'GET',
    url: 'https://api.demo.reqly.dev/v1/charges/ch_1',
    status: 200,
    durationMs: 178,
    size: 2140,
    env: 'dev',
    createdAt: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
    reqHeaders: { accept: ['application/json'] },
    reqBody: '',
    respHeaders: { 'content-type': ['application/json'] },
    respBody: '{\n  "id": "ch_1",\n  "amount": 1999,\n  "status": "succeeded"\n}',
  },
]

const historyDetail = (id: string): HistoryDetail =>
  historyEntries.find((h) => h.id === id) ?? historyEntries[0]

export const demoHistoryAdapter: HistoryAdapter = {
  list: async (limit) => historyEntries.slice(0, limit),
  show: async (id) => historyDetail(id),
  search: async (q) =>
    historyEntries.filter((h) => h.url.includes(q)),
  clear: async () => {},
  replay: async () => null,
  listCookies: async () => [],
  deleteCookie: async () => {},
  clearCookies: async () => {},
}

export const demoBootstrapAdapter: WorkspaceBootstrapAdapter = {
  status: async () => ({ found: true, path: tree.path }),
  restoreLast: async () => ({ found: true, path: tree.path }),
  pickFolder: async () => '',
  open: async () => {},
  create: async () => {},
}
