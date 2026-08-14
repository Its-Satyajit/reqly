import { WorkspaceSidebar } from '../components'
import { RequestEditor } from '../features/request-editor/RequestEditor'
import { ResponseViewer } from '../features/response-viewer/ResponseViewer'
import { useWorkspaceStore } from '../stores'

export function App() {
  const environments = useWorkspaceStore((s) => s.environments)
  const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId)
  const setActiveEnvironment = useWorkspaceStore((s) => s.setActiveEnvironment)

  const activeEnvironment = environments.find((e) => e.id === activeEnvironmentId)

  return (
    <div className="flex h-screen flex-col">
      <header className="flex h-12 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-sm font-semibold">API Client</h1>
        <select
          value={activeEnvironment?.id ?? ''}
          onChange={(e) => setActiveEnvironment(e.target.value || null)}
          className="rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground"
        >
          <option value="">No environment</option>
          {environments.map((env) => (
            <option key={env.id} value={env.id}>
              {env.name}
            </option>
          ))}
        </select>
      </header>
      <div className="flex min-h-0 flex-1">
        <WorkspaceSidebar />
        <main className="flex min-w-0 flex-1 flex-col lg:flex-row">
          <section className="min-h-0 flex-1 border-r border-border lg:w-1/2">
            <RequestEditor />
          </section>
          <section className="min-h-0 flex-1 lg:w-1/2">
            <ResponseViewer />
          </section>
        </main>
      </div>
    </div>
  )
}