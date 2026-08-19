import { useEffect, useState } from 'react'
import { useWorkspaceStore } from '../../stores'
import { Button } from '../../components'
import { EnvironmentEditor } from './EnvironmentEditor'
import { SecretsEditor } from './SecretsEditor'

const inputClass =
  'rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus-visible:border-ring'

/**
 * EnvironmentsView manages the workspace's environments: list, create, set
 * the active one, and edit existing environments (description + variables)
 * through an in-memory editor with explicit Save.
 */
export function EnvironmentsView() {
  const environments = useWorkspaceStore((s) => s.environments)
  const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId)
  const environmentsError = useWorkspaceStore((s) => s.environmentsError)
  const envAdapter = useWorkspaceStore((s) => s.envAdapter)
  const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments)
  const setHasUnsavedEnvChanges = useWorkspaceStore((s) => s.setHasUnsavedEnvChanges)

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [setActiveError, setSetActiveError] = useState<string | null>(null)
  const [editingName, setEditingName] = useState<string | null>(null)

  useEffect(() => {
    void refreshEnvironments()
  }, [refreshEnvironments])

  const editing = environments.find((e) => e.name === editingName) ?? null

  const onEdit = (name: string) => {
    if (editingName === name) {
      setEditingName(null)
      return
    }
    setEditingName(name)
  }

  const onCreate = async () => {
    setCreateError(null)
    const trimmed = name.trim()
    if (!trimmed) {
      setCreateError('Name the environment — e.g. "dev" or "staging".')
      return
    }
    setCreating(true)
    try {
      await envAdapter.create(trimmed, description.trim(), {})
      setName('')
      setDescription('')
      await refreshEnvironments()
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : String(err))
    } finally {
      setCreating(false)
    }
  }

  const onSetActive = async (name: string) => {
    setSetActiveError(null)
    try {
      await envAdapter.setActive(name)
      await refreshEnvironments()
    } catch (err) {
      setSetActiveError(err instanceof Error ? err.message : String(err))
    }
  }

  const onDelete = async (name: string) => {
    if (!window.confirm(`Delete environment "${name}"? This removes its file and its secrets.`)) {
      return
    }
    setSetActiveError(null)
    try {
      await envAdapter.delete(name)
      if (editingName === name) setEditingName(null)
      await refreshEnvironments()
    } catch (err) {
      setSetActiveError(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-5 p-6">
      <div>
        <h2 className="text-sm font-semibold">Environments</h2>
        <p className="text-xs text-muted-foreground">
          Named sets of variables (and secrets) applied to requests. The active
          environment is shared with the CLI.
        </p>
      </div>

      <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          New environment
        </p>
        <div className="flex flex-col gap-2 sm:flex-row">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="name (e.g. dev)"
            spellCheck={false}
            className={`${inputClass} font-mono`}
          />
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="description (optional)"
            className={`${inputClass} flex-1`}
          />
          <Button onClick={() => void onCreate()} disabled={creating || !name.trim()}>
            Create
          </Button>
        </div>
        {createError && <p className="text-xs text-destructive">{createError}</p>}
      </div>

      {setActiveError && <p className="text-xs text-destructive">{setActiveError}</p>}

      {editing && (
        <div className="flex flex-col gap-3">
          <EnvironmentEditor
            env={editing}
            onCancel={() => {
              setHasUnsavedEnvChanges(false)
              setEditingName(null)
            }}
          />
          <SecretsEditor envName={editing.name} secretNames={editing.secrets} />
        </div>
      )}

      {environmentsError ? (
        <p className="rounded-md border border-border bg-card p-3 text-xs text-destructive">
          {environmentsError}
        </p>
      ) : environments.length === 0 ? (
        <div className="rounded-md border border-dashed border-border bg-card p-4 text-center">
          <p className="text-xs text-muted-foreground">
            No environments yet. Create one above to start modeling your targets.
          </p>
        </div>
      ) : (
        <ul className="flex flex-col gap-2">
          {environments.map((env) => {
            const active = env.id === activeEnvironmentId
            return (
              <li
                key={env.id}
                className="flex items-center justify-between gap-3 rounded-md border border-border bg-card px-3 py-2"
              >
                <button
                  type="button"
                  onClick={() => onEdit(env.name)}
                  className="flex min-w-0 flex-1 flex-col gap-0.5 text-left"
                >
                  <span className="flex items-center gap-2 text-sm font-medium">
                    <span className="font-mono">{env.name}</span>
                    {active && (
                      <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                        active
                      </span>
                    )}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {env.description || `${Object.keys(env.variables).length} variable(s)`}
                  </span>
                </button>
                <div className="flex items-center gap-2">
                  {editingName === env.name && (
                    <Button variant="outline" size="sm" onClick={() => onEdit(env.name)}>
                      Editing…
                    </Button>
                  )}
                  {!active && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => void onSetActive(env.name)}
                    >
                      Use
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-destructive hover:bg-destructive/10"
                    onClick={() => void onDelete(env.name)}
                  >
                    Delete
                  </Button>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}